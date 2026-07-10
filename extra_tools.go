package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// ─────────────────────────────────────────
// QR CODE GENERATOR
// .qr <text>  → QR image generate karta hai
// ─────────────────────────────────────────
func handleQR(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.qr <text or URL>`\nExample: `.qr https://wa.me/923001234567`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔲")

	apiURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=512x512&data=%s", url.QueryEscape(args))
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		replyMessage(client, v, "❌ QR generation failed. Try again.")
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	up, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		replyMessage(client, v, "❌ Upload failed.")
		return
	}

	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL: proto.String(up.URL), DirectPath: proto.String(up.DirectPath),
			MediaKey: up.MediaKey, Mimetype: proto.String("image/png"),
			Caption:       proto.String(fmt.Sprintf("🔲 *QR Code Generated!*\n📝 *Text:* %s", args)),
			FileEncSHA256: up.FileEncSHA256, FileSHA256: up.FileSHA256,
			FileLength: proto.Uint64(uint64(len(data))),
		},
	})
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// OCR — IMAGE TO TEXT
// Reply to an image with .ocr
// ─────────────────────────────────────────
func handleOCR(client *whatsmeow.Client, v *events.Message) {
	extMsg := v.Message.GetExtendedTextMessage()
	if extMsg == nil || extMsg.ContextInfo == nil || extMsg.ContextInfo.QuotedMessage == nil {
		replyMessage(client, v, "❌ Please *reply to an image* with `.ocr`")
		return
	}

	img := extMsg.ContextInfo.QuotedMessage.GetImageMessage()
	if img == nil {
		replyMessage(client, v, "❌ Replied message must be an *image*.")
		return
	}

	react(client, v.Info.Chat, v.Info.ID, "🔍")
	imgData, err := client.Download(context.Background(), img)
	if err != nil {
		replyMessage(client, v, "❌ Failed to download image.")
		return
	}

	// OCR.space free API
	b64 := base64.StdEncoding.EncodeToString(imgData)
	formData := url.Values{
		"base64Image": {"data:image/jpeg;base64," + b64},
		"language":    {"eng"},
		"isOverlayRequired": {"false"},
	}

	req, _ := http.NewRequest("POST", "https://api.ocr.space/parse/image", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", "helloworld") // free public key

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		replyMessage(client, v, "❌ OCR server error.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		ParsedResults []struct {
			ParsedText string `json:"ParsedText"`
		} `json:"ParsedResults"`
		IsErroredOnProcessing bool `json:"IsErroredOnProcessing"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.IsErroredOnProcessing || len(result.ParsedResults) == 0 || strings.TrimSpace(result.ParsedResults[0].ParsedText) == "" {
		replyMessage(client, v, "❌ No text found in the image.")
		return
	}

	text := strings.TrimSpace(result.ParsedResults[0].ParsedText)
	replyMessage(client, v, fmt.Sprintf("📝 *OCR Result:*\n\n%s", text))
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// TTP — TEXT TO PICTURE STICKER
// .ttp <text>  → text wala sticker
// ─────────────────────────────────────────
func handleTTP(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.ttp <your text>`\nExample: `.ttp Bunny MD 🐰`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🎨")

	// Use textcraft API for stylish text image
	apiURL := fmt.Sprintf("https://api.textcraft.net/image.php?text=%s&font=impact&color=ffffff&size=120&back=000000", url.QueryEscape(args))
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		// Fallback: flamingtext style
		apiURL = fmt.Sprintf("https://api.multiavatar.com/%s.png", url.QueryEscape(args))
		resp, err = http.Get(apiURL)
		if err != nil {
			replyMessage(client, v, "❌ TTP service unavailable. Try again later.")
			return
		}
	}
	defer resp.Body.Close()

	imgData, _ := io.ReadAll(resp.Body)
	if len(imgData) < 500 {
		replyMessage(client, v, "❌ Failed to generate image. Try a shorter text.")
		return
	}

	// Upload as WebP sticker
	up, err := client.Upload(context.Background(), imgData, whatsmeow.MediaImage)
	if err != nil {
		replyMessage(client, v, "❌ Upload failed.")
		return
	}

	// Send as sticker-captioned image (WhatsApp sticker needs exif; send as image with caption)
	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL: proto.String(up.URL), DirectPath: proto.String(up.DirectPath),
			MediaKey: up.MediaKey, Mimetype: proto.String("image/png"),
			Caption:       proto.String("🎨 *Text Sticker!* Convert to sticker by forwarding & saving."),
			FileEncSHA256: up.FileEncSHA256, FileSHA256: up.FileSHA256,
			FileLength: proto.Uint64(uint64(len(imgData))),
		},
	})
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// BASE64 ENCODE / DECODE
// .encode <text>   → base64 encode
// .decode <text>   → base64 decode
// ─────────────────────────────────────────
func handleEncode(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.encode <text>`")
		return
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(args))
	replyMessage(client, v, fmt.Sprintf("🔐 *Base64 Encoded:*\n\n`%s`", encoded))
}

func handleDecode(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.decode <base64 text>`")
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(args))
	if err != nil {
		replyMessage(client, v, "❌ Invalid Base64 string.")
		return
	}
	replyMessage(client, v, fmt.Sprintf("🔓 *Base64 Decoded:*\n\n%s", string(decoded)))
}

// ─────────────────────────────────────────
// DICE & COIN FLIP
// .dice        → 1-6 random number
// .coin        → Heads ya Tails
// ─────────────────────────────────────────
func handleDice(client *whatsmeow.Client, v *events.Message) {
	n := rand.Intn(6) + 1
	faces := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣"}
	replyMessage(client, v, fmt.Sprintf("🎲 *Dice Roll!*\n\nResult: %s *(%d)*", faces[n-1], n))
}

func handleCoin(client *whatsmeow.Client, v *events.Message) {
	if rand.Intn(2) == 0 {
		replyMessage(client, v, "🪙 *Coin Flip!*\n\nResult: 👑 *HEADS!*")
	} else {
		replyMessage(client, v, "🪙 *Coin Flip!*\n\nResult: 🔵 *TAILS!*")
	}
}

// ─────────────────────────────────────────
// WIKIPEDIA SEARCH
// .wiki <query>
// ─────────────────────────────────────────
func handleWiki(client *whatsmeow.Client, v *events.Message, query string) {
	if query == "" {
		replyMessage(client, v, "❌ *Usage:* `.wiki <topic>`\nExample: `.wiki Elon Musk`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "📚")

	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(strings.ReplaceAll(query, " ", "_")))
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		replyMessage(client, v, "❌ Wikipedia article not found. Try a different search term.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Title   string `json:"title"`
		Extract string `json:"extract"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Extract == "" {
		replyMessage(client, v, "❌ No Wikipedia article found.")
		return
	}

	// Truncate to 800 chars
	extract := result.Extract
	if len(extract) > 800 {
		extract = extract[:800] + "..."
	}

	text := fmt.Sprintf("📚 *%s*\n\n%s\n\n🔗 %s", result.Title, extract, result.ContentURLs.Desktop.Page)
	replyMessage(client, v, text)
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// BROADCAST
// .bc <message>  → sab groups mein message bhejta hai (owner only)
// ─────────────────────────────────────────
func handleBroadcast(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.bc <your message>`\n\n⚠️ This will send to ALL groups the bot is in!")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "📢")

	groups, err := client.GetJoinedGroups(context.Background())
	if err != nil || len(groups) == 0 {
		replyMessage(client, v, "❌ Failed to fetch groups or bot is in no groups.")
		return
	}

	replyMessage(client, v, fmt.Sprintf("📢 *Broadcasting to %d groups...*\nPlease wait.", len(groups)))

	sent := 0
	failed := 0
	bcMsg := fmt.Sprintf("📢 *BROADCAST MESSAGE*\n\n%s\n\n_Sent by Bunny MD_", args)

	for _, g := range groups {
		_, err := client.SendMessage(context.Background(), g.JID, &waProto.Message{
			Conversation: proto.String(bcMsg),
		})
		if err != nil {
			failed++
		} else {
			sent++
		}
		time.Sleep(1 * time.Second) // flood protection
	}

	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *Broadcast Complete!*\n\n📤 Sent: *%d*\n❌ Failed: *%d*", sent, failed))
}

// ─────────────────────────────────────────
// RANK / XP SYSTEM
// .rank  → apna rank dekho
// .topleaderboard → top 10 users
// XP auto-deta hai har message pe
// ─────────────────────────────────────────

type UserXP struct {
	XP    int
	Level int
	Name  string
}

var xpData = make(map[string]*UserXP)
var xpMu sync.Mutex
var xpCooldown = make(map[string]time.Time)

func addXP(userJID string, name string) {
	xpMu.Lock()
	defer xpMu.Unlock()

	// 30 second cooldown per user
	if last, ok := xpCooldown[userJID]; ok && time.Since(last) < 30*time.Second {
		return
	}
	xpCooldown[userJID] = time.Now()

	if _, ok := xpData[userJID]; !ok {
		xpData[userJID] = &UserXP{Name: name}
	}
	xpData[userJID].XP += rand.Intn(10) + 5 // 5-15 XP per message
	xpData[userJID].Name = name

	// Level up: every 100 XP = 1 level
	xpData[userJID].Level = xpData[userJID].XP / 100
}

func handleRank(client *whatsmeow.Client, v *events.Message) {
	xpMu.Lock()
	defer xpMu.Unlock()

	userKey := v.Info.Sender.User
	user, ok := xpData[userKey]
	if !ok {
		replyMessage(client, v, "ℹ️ You have no XP yet! Start chatting to earn XP. 💬")
		return
	}

	nextLevel := (user.Level + 1) * 100
	progress := user.XP % 100

	bar := ""
	for i := 0; i < 10; i++ {
		if i < progress/10 {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	text := fmt.Sprintf(`🏆 *RANK CARD*

👤 *User:* @%s
⭐ *Level:* %d
💎 *XP:* %d / %d
📊 *Progress:* [%s]

_Keep chatting to level up!_ 🚀`, userKey, user.Level, user.XP, nextLevel, bar)

	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: []string{v.Info.Sender.String()},
			},
		},
	})
}

func handleLeaderboard(client *whatsmeow.Client, v *events.Message) {
	xpMu.Lock()
	defer xpMu.Unlock()

	if len(xpData) == 0 {
		replyMessage(client, v, "ℹ️ No XP data yet! Start chatting to earn XP. 💬")
		return
	}

	// Sort by XP
	type entry struct {
		jid  string
		data *UserXP
	}
	var sorted []entry
	for k, d := range xpData {
		sorted = append(sorted, entry{k, d})
	}
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].data.XP > sorted[i].data.XP {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	medals := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	var sb strings.Builder
	sb.WriteString("🏆 *XP LEADERBOARD — TOP 10*\n\n")

	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		e := sorted[i]
		sb.WriteString(fmt.Sprintf("%s @%s\n   ⭐ Lvl %d | 💎 %d XP\n\n", medals[i], e.jid, e.data.Level, e.data.XP))
	}

	var mentions []string
	for i := 0; i < limit; i++ {
		mentions = append(mentions, sorted[i].jid+"@s.whatsapp.net")
	}

	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(sb.String()),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: mentions,
			},
		},
	})
}

// ─────────────────────────────────────────
// RANDOM FACT
// .fact → ek random fun fact bhejta hai
// ─────────────────────────────────────────
func handleFact(client *whatsmeow.Client, v *events.Message) {
	react(client, v.Info.Chat, v.Info.ID, "🧠")
	resp, err := http.Get("https://uselessfacts.jsph.pl/api/v2/facts/random?language=en")
	if err != nil {
		replyMessage(client, v, "❌ Failed to fetch fact. Try again.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Text == "" {
		replyMessage(client, v, "❌ No fact found. Try again.")
		return
	}
	replyMessage(client, v, fmt.Sprintf("🧠 *Random Fact:*\n\n%s", result.Text))
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// JOKE
// .joke → ek random joke bhejta hai
// ─────────────────────────────────────────
func handleJoke(client *whatsmeow.Client, v *events.Message) {
	react(client, v.Info.Chat, v.Info.ID, "😂")
	resp, err := http.Get("https://official-joke-api.appspot.com/random_joke")
	if err != nil {
		replyMessage(client, v, "❌ Failed to fetch joke. Try again.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Setup     string `json:"setup"`
		Punchline string `json:"punchline"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Setup == "" {
		replyMessage(client, v, "❌ No joke found. Try again.")
		return
	}
	replyMessage(client, v, fmt.Sprintf("😂 *Random Joke:*\n\n❓ %s\n\n😄 %s", result.Setup, result.Punchline))
	react(client, v.Info.Chat, v.Info.ID, "✅")
}

// ─────────────────────────────────────────
// CALCULATE
// .calc <expression>  → math solve karta hai
// ─────────────────────────────────────────
func handleCalc(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.calc <math expression>`\nExample: `.calc 25 * 4 + 10`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔢")

	// Use mathjs API
	apiURL := fmt.Sprintf("https://api.mathjs.org/v4/?expr=%s", url.QueryEscape(args))
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		replyMessage(client, v, "❌ Calculator error. Check your expression.")
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	resultStr := strings.TrimSpace(string(result))

	if strings.Contains(resultStr, "Error") || resultStr == "" {
		replyMessage(client, v, "❌ Invalid expression. Example: `.calc 2+2`")
		return
	}

	replyMessage(client, v, fmt.Sprintf("🔢 *Calculator*\n\n📝 *Expression:* `%s`\n✅ *Result:* `%s`", args, resultStr))
	react(client, v.Info.Chat, v.Info.ID, "✅")
}
