package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"strings"


	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)




func handleScreenshot(client *whatsmeow.Client, v *events.Message, targetUrl string) {
	if targetUrl == "" {
		replyMessage(client, v, "⚠️ *Usage:* `.ss https://example.com`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "📸")

	
	apiURL := fmt.Sprintf("https://api.screenshotmachine.com/?key=54be93&device=phone&dimension=1290x2796&url=%s", url.QueryEscape(targetUrl))

	
	resp, err := http.Get(apiURL)
	if err != nil {
		replyMessage(client, v, "❌ Screenshot engine failed to connect.")
		return
	}
	defer resp.Body.Close()

	
	fileData, err := io.ReadAll(resp.Body)
	if err != nil || len(fileData) < 1000 { 
		replyMessage(client, v, "❌ Failed to capture. Website might be blocking screenshots.")
		return 
	}

	
	up, err := client.Upload(context.Background(), fileData, whatsmeow.MediaImage)
	if err != nil {
		replyMessage(client, v, "❌ WhatsApp rejected the media upload.")
		return
	}

	
	finalMsg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:        proto.String(up.URL),
			DirectPath: proto.String(up.DirectPath),
			MediaKey:   up.MediaKey,
			Mimetype:   proto.String("image/jpeg"),
			Caption:    proto.String("✅ *Web Capture Success*\n🌐 " + targetUrl),
			FileSHA256: up.FileSHA256,
			FileEncSHA256: up.FileEncSHA256,
			FileLength: proto.Uint64(uint64(len(fileData))),
		},
	}

	client.SendMessage(context.Background(), v.Info.Chat, finalMsg)
	react(client, v.Info.Chat, v.Info.ID, "✅")
}




func handleTranslate(client *whatsmeow.Client, v *events.Message, args string) {
	parts := strings.Fields(args)
	if len(parts) == 0 {
		replyMessage(client, v, "❌ *Usage:* `.tr urdu Text` or reply to a message with `.tr urdu`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔄")

	
	langMap := map[string]string{
		"urdu": "ur", "ur": "ur", "english": "en", "en": "en",
		"arabic": "ar", "ar": "ar", "hindi": "hi", "hi": "hi",
		"pashto": "ps", "ps": "ps", "punjabi": "pa", "pa": "pa",
	}
	targetLang := langMap[strings.ToLower(parts[0])]
	if targetLang == "" { targetLang = "ur" } 

	
	var textToTranslate string
	extMsg := v.Message.GetExtendedTextMessage()
	if extMsg != nil && extMsg.ContextInfo != nil && extMsg.ContextInfo.QuotedMessage != nil {
		q := extMsg.ContextInfo.QuotedMessage
		if q.Conversation != nil {
			textToTranslate = *q.Conversation
		} else if q.ExtendedTextMessage != nil {
			textToTranslate = *q.ExtendedTextMessage.Text
		}
	}

	if textToTranslate == "" && len(parts) > 1 {
		textToTranslate = strings.Join(parts[1:], " ")
	}

	if textToTranslate == "" {
		replyMessage(client, v, "❌ Please provide text or reply to a message to translate.")
		return
	}

	
	apiURL := fmt.Sprintf("https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s", targetLang, url.QueryEscape(textToTranslate))
	resp, err := http.Get(apiURL)
	if err != nil {
		replyMessage(client, v, "❌ Translation Server Error.")
		return
	}
	defer resp.Body.Close()

	var result []interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	
	translatedText := ""
	if len(result) > 0 {
		innerArray, ok := result[0].([]interface{})
		if ok {
			for _, item := range innerArray {
				strArray, ok2 := item.([]interface{})
				if ok2 && len(strArray) > 0 {
					translatedText += fmt.Sprintf("%v", strArray[0])
				}
			}
		}
	}

	if translatedText != "" {
		replyMessage(client, v, fmt.Sprintf("🌐 *Translation (%s):*\n\n%s", strings.ToUpper(targetLang), translatedText))
		react(client, v.Info.Chat, v.Info.ID, "✅")
	} else {
		replyMessage(client, v, "❌ Failed to parse translation.")
	}
}




func handleImageGen(client *whatsmeow.Client, v *events.Message, prompt string) {
	if prompt == "" {
		replyMessage(client, v, "❌ *Usage:* `.img A futuristic cyber city at night`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🎨")

	
	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=1024&height=1024&nologo=true", url.PathEscape(prompt))
	
	resp, err := http.Get(apiURL)
	if err != nil {
		replyMessage(client, v, "❌ Failed to connect to AI Engine.")
		return
	}
	defer resp.Body.Close()

	fileData, _ := io.ReadAll(resp.Body)
	up, err := client.Upload(context.Background(), fileData, whatsmeow.MediaImage)
	if err != nil { return }

	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL: proto.String(up.URL), DirectPath: proto.String(up.DirectPath),
			MediaKey: up.MediaKey, Mimetype: proto.String("image/jpeg"),
			Caption: proto.String("✨ *AI Generation Complete*\n🎨 *Prompt:* " + prompt),
			FileEncSHA256: up.FileEncSHA256, FileSHA256: up.FileSHA256,
			FileLength: proto.Uint64(uint64(len(fileData))),
		},
	})
	react(client, v.Info.Chat, v.Info.ID, "✅")
}




func handleWeather(client *whatsmeow.Client, v *events.Message, city string) {
	if city == "" { city = "Faisalabad" } 
	react(client, v.Info.Chat, v.Info.ID, "☁️")

	
	apiURL := fmt.Sprintf("https://wttr.in/%s?format=%%l:+%%C+%%c+%%t+(Feels+like+%%f)\\nWind:+%%w\\nHumidity:+%%h", url.PathEscape(city))
	
	resp, err := http.Get(apiURL)
	if err != nil { return }
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	replyMessage(client, v, fmt.Sprintf("🌤️ *WEATHER REPORT*\n\n%s\n\n_Powered by Bunny MD_", string(data)))
	react(client, v.Info.Chat, v.Info.ID, "✅")
}




func handleGoogle(client *whatsmeow.Client, v *events.Message, query string) {
	if query == "" {
		replyMessage(client, v, "❌ *Usage:* `.google Bunny MD`")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔍")

	searchLink := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))
	
	msg := fmt.Sprintf("🔍 *GOOGLE SEARCH*\n\n💬 *Query:* %s\n\n🔗 *Click here for results:*\n%s", query, searchLink)
	replyMessage(client, v, msg)
	react(client, v.Info.Chat, v.Info.ID, "✅")
}
