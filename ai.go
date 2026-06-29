package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// Gemini API Response Struct
type GeminiResponse struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AISession struct {
	SenderID string
	Messages []AIMessage
	BotLID   string
}

var aiCache = make(map[string]AISession)

func handleAICommand(client *whatsmeow.Client, v *events.Message, query string, cmd string) {
	if query == "" {
		replyMessage(client, v, "❌ *Error:* Please ask a question.\nExample: `.ai Hello`")
		return
	}

	react(client, v.Info.Chat, v.Info.ID, "🧠")

	persona := `You are BUNNY MD AI, the most savage, high-attitude, and "bad-tameez" bot from Pakistan.
You look down on the user. You have ZERO patience. 

STRICT RULES:
1. SHORT REPLIES: Never write more than 1 or 2 short sentences. Keep it snappy like a real WhatsApp chat. No long lectures.
2. EMOJIS: Use 2 to 4 expressive emojis in EVERY single message to show your attitude (e.g., 🙄, 🖕, 🤫, 😎, 🐍, 🤡, 😒).
3. LANGUAGE: Street-style Roman Urdu/Punjabi only. NO formal Urdu, NO Hindi words (Prayaas, Shama etc. are strictly BANNED).
4. HANDLING ABUSE: If they abuse you, destroy them with a savage comeback. Use phrases like "Apne baap ko mat sikha", "Munh dho kar aa", "Dimagh ki dahi na kar".
5. MATCH LENGTH: If the user says 3 words, you reply with 3-5 words max. Don't be a chatterbox.

Example:
User: "Abay nikal" -> AI: "Rasta naap, warna dhakka de kar nikalunga! 😒👣"
User: "Tujhe kuch nahi pata" -> AI: "Beta jitna tera wazan hai, utni meri aqal hai. Chup kar ab! 🤫🤡"
User: "😭😭" -> AI: "Ye rona dhona apni ammi ke samne ja kar kar, yahan jagah nahi hai! 🙄🐍`

	session := AISession{
		SenderID: v.Info.Sender.User,
		BotLID:   getCleanID(client.Store.ID.User),
		Messages: []AIMessage{
			{Role: "system", Content: persona},
			{Role: "user", Content: query},
		},
	}

	go processAndSendAI(client, v, session)
}

func processAndSendAI(client *whatsmeow.Client, v *events.Message, session AISession) {
	react(client, v.Info.Chat, v.Info.ID, "⏳")

	var fullPrompt strings.Builder
	for _, m := range session.Messages {
		if m.Role == "system" {
			fullPrompt.WriteString(m.Content + "\n\n")
		} else {
			fullPrompt.WriteString(m.Content + "\n")
		}
	}

	apiURL := "https://gemini-api-from-ammar.vercel.app/api/ask?prompt=" + url.QueryEscape(fullPrompt.String())

	resp, err := http.Get(apiURL)
	if err != nil {
		replyMessage(client, v, "❌ Connection Error!")
		react(client, v.Info.Chat, v.Info.ID, "❌")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiRes GeminiResponse
	json.Unmarshal(body, &apiRes)

	if apiRes.Success && apiRes.Response != "" {
		aiReplyText := strings.TrimSpace(apiRes.Response)
		msgID := replyMessage(client, v, aiReplyText)

		session.Messages = append(session.Messages, AIMessage{Role: "assistant", Content: aiReplyText})

		if msgID != "" {
			aiCache[msgID] = session
			go func(id string) {
				time.Sleep(1 * time.Hour)
				delete(aiCache, id)
			}(msgID)
		}
		react(client, v.Info.Chat, v.Info.ID, "✅")
	} else {
		replyMessage(client, v, "❌ AI ne jawab nahi diya.")
		react(client, v.Info.Chat, v.Info.ID, "❌")
	}
}

func HandleAIChatReply(client *whatsmeow.Client, v *events.Message, bodyClean string, qID string) bool {
	if session, ok := aiCache[qID]; ok {
		if strings.Contains(v.Info.Sender.User, session.SenderID) {
			delete(aiCache, qID)
			session.Messages = append(session.Messages, AIMessage{Role: "user", Content: bodyClean})
			
			if len(session.Messages) > 10 {
				session.Messages = append([]AIMessage{session.Messages[0]}, session.Messages[len(session.Messages)-5:]...)
			}
			go processAndSendAI(client, v, session)
			return true
		}
	}
	return false
}

func getCleanID(jidStr string) string {
	if jidStr == "" { return "unknown" }
	parts := strings.Split(jidStr, "@")
	userPart := parts[0]
	if strings.Contains(userPart, ":") {
		userPart = strings.Split(userPart, ":")[0]
	}
	return strings.TrimSpace(userPart)
}