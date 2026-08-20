const express = require("express");
const path = require("path");
const crypto = require("crypto");

const app = express();
app.use(express.json());

// ===== Security headers =====
app.use((req, res, next) => {
  res.setHeader("X-Frame-Options", "DENY");
  res.setHeader("X-Content-Type-Options", "nosniff");
  res.setHeader("Referrer-Policy", "strict-origin-when-cross-origin");
  res.setHeader("Permissions-Policy", "camera=(), microphone=(), geolocation=()");
  next();
});

app.use(express.static(path.join(__dirname, "public")));

// =========================================================
// ROTATING TOKEN — HMAC-signed, tied to a 5-minute time window.
// No database or cron job needed; validity is derived purely from
// (time window + secret). The secret never leaves this file.
// =========================================================
const TOKEN_LIFETIME_MS = 5 * 60 * 1000;

function getSecret() {
  const secret = process.env.NETWORK_TOKEN_SECRET;
  if (!secret) throw new Error("NETWORK_TOKEN_SECRET is not set");
  return secret;
}
function currentWindow() {
  return Math.floor(Date.now() / TOKEN_LIFETIME_MS);
}
function signToken(scope, window) {
  return crypto.createHmac("sha256", getSecret()).update(`${scope}:${window}`).digest("hex");
}
function generateToken(scope) {
  const window = currentWindow();
  const signature = signToken(scope, window);
  return { token: `${window}.${signature}`, expiresAt: (window + 1) * TOKEN_LIFETIME_MS };
}
function verifyToken(scope, tokenStr) {
  if (!tokenStr || !tokenStr.includes(".")) return false;
  const [windowStr, signature] = tokenStr.split(".");
  const window = Number(windowStr);
  if (!Number.isFinite(window)) return false;
  const cur = currentWindow();
  if (window !== cur && window !== cur - 1) return false;
  const expected = signToken(scope, window);
  try {
    return crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(signature));
  } catch {
    return false;
  }
}

// =========================================================
// RATE LIMITER — simple in-memory
// =========================================================
const buckets = new Map();
function checkRateLimit(key, maxRequests, windowMs) {
  const now = Date.now();
  const existing = buckets.get(key);
  if (!existing || now > existing.resetAt) {
    buckets.set(key, { count: 1, resetAt: now + windowMs });
    return true;
  }
  if (existing.count >= maxRequests) return false;
  existing.count += 1;
  return true;
}

// ===== Token issuing =====
const ALLOWED_SCOPES = new Set(["network-check"]);

app.get("/api/token", (req, res) => {
  const scope = req.query.scope;
  if (!ALLOWED_SCOPES.has(scope)) {
    return res.status(400).json({ error: "Invalid scope" });
  }
  const { token, expiresAt } = generateToken(scope);
  res.json({ token, expiresAt });
});

// =========================================================
// NETWORK CHECK
// Calls the real lookup API (which returns name/CNIC/address alongside
// the network), but discards everything except `network` before
// anything is sent back to the browser or logged. No request or
// response data from this endpoint is written to logs.
// =========================================================
const UPSTREAM_API = "https://wasifali.biz.id/public_apis/sim-info-api.php";

function normalizeNumber(raw) {
  const digits = String(raw || "").replace(/\D/g, "");
  let local = digits;
  if (local.startsWith("92")) local = local.slice(2);
  if (local.startsWith("0")) local = local.slice(1);
  return local;
}

app.post("/api/network-check", async (req, res) => {
  const ip = req.headers["x-forwarded-for"]?.split(",")[0]?.trim() || req.ip;
  if (!checkRateLimit(`network-check:${ip}`, 20, 60_000)) {
    return res.status(429).json({ error: "Too many requests. Please slow down." });
  }

  const { number, token } = req.body || {};
  if (!number || !token) {
    return res.status(400).json({ error: "Missing number or token" });
  }
  if (!verifyToken("network-check", token)) {
    return res.status(401).json({ error: "Invalid or expired token. Refresh the page and try again." });
  }

  const local = normalizeNumber(number);
  if (local.length !== 10 || !local.startsWith("3")) {
    return res.json({ carrier: "Unknown", valid: false });
  }

  try {
    const upstreamRes = await fetch(`${UPSTREAM_API}?search=${encodeURIComponent("0" + local)}`);
    if (!upstreamRes.ok) {
      return res.status(502).json({ error: "Lookup service unavailable right now." });
    }
    const data = await upstreamRes.json();
    const record = data?.records?.[0];

    // Only the network name ever leaves this function. name/cnic/address
    // in `record` are discarded here and never sent or stored anywhere.
    const network = record?.network || null;

    if (!network) return res.json({ carrier: "Unknown", valid: false });
    return res.json({ carrier: network, valid: true });
  } catch {
    return res.status(500).json({ error: "Something went wrong. Please try again." });
  }
});

const PORT = process.env.PORT || 8080;
if (!process.env.NETWORK_TOKEN_SECRET) {
  console.error("NETWORK_TOKEN_SECRET environment variable is not set. Refusing to start.");
  process.exit(1);
}

app.listen(PORT, () => {
  console.log(`BUNNY Network Checker server running on :${PORT}`);
});
