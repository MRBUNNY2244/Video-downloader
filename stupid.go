package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Helper: Safely extracts ContextInfo (Unique to avoid name collisions)
func getStupidContextInfo(msg *waE2E.Message) *waE2E.ContextInfo {
	if msg == nil {
		return nil
	}
	if msg.ExtendedTextMessage != nil {
		return msg.ExtendedTextMessage.ContextInfo
	}
	if msg.ImageMessage != nil {
		return msg.ImageMessage.ContextInfo
	}
	if msg.VideoMessage != nil {
		return msg.VideoMessage.ContextInfo
	}
	if msg.DocumentMessage != nil {
		return msg.DocumentMessage.ContextInfo
	}
	if msg.AudioMessage != nil {
		return msg.AudioMessage.ContextInfo
	}
	return nil
}

// Helper: Determine target JID (who) exactly like JS logic
func getStupidTargetJID(v *events.Message) types.JID {
	ctxInfo := getStupidContextInfo(v.Message)
	if ctxInfo != nil {
		// quotedMsg ? quotedMsg.sender
		if ctxInfo.Participant != nil {
			parsed, err := types.ParseJID(*ctxInfo.Participant)
			if err == nil {
				return parsed
			}
		}
		// mentionedJid && mentionedJid[0] ? mentionedJid[0]
		if len(ctxInfo.MentionedJID) > 0 {
			parsed, err := types.ParseJID(ctxInfo.MentionedJID[0])
			if err == nil {
				return parsed
			}
		}
	}
	// : sender
	return v.Info.Sender
}

// Handler: .stupid Command
func handleStupidCommand(client *whatsmeow.Client, v *events.Message, fullArgs string) {
	// Entire outer try-catch block simulation
	err := processStupidLogic(client, v, fullArgs)
	if err != nil {
		// On any catch error, send exact string requested
		replyMessage(client, v, "❌ Sorry, I couldn't generate the stupid card. Please try again later!")
	}
}

// Core execution logic (Simulates the "try" block)
func processStupidLogic(client *whatsmeow.Client, v *events.Message, fullArgs string) error {
	// let who = quotedMsg ? quotedMsg.sender : mentionedJid && mentionedJid[0] ? ...
	who := getStupidTargetJID(v)

	// let text = args && args.length > 0 ? args.join(' ') : 'im+stupid';
	text := strings.TrimSpace(fullArgs)
	if text == "" {
		text = "im stupid"
	}

	// let avatarUrl;
	avatarUrl := "https://telegra.ph/file/24fa902ead26340f3df2c.png"
	
	// try { avatarUrl = await sock.profilePictureUrl(who, 'image'); } catch (error) { ... }
	ppInfo, err := client.GetProfilePictureInfo(context.Background(), who, &whatsmeow.GetProfilePictureParams{
		Preview: false,
	})
	if err == nil && ppInfo != nil && ppInfo.URL != "" {
		avatarUrl = ppInfo.URL
	}

	// const apiUrl = `...its-so-stupid?avatar=${encodeURIComponent(avatarUrl)}&dog=${encodeURIComponent(text)}`
	apiUrl := fmt.Sprintf("https://some-random-api.com/canvas/misc/its-so-stupid?avatar=%s&dog=%s",
		url.QueryEscape(avatarUrl),
		url.QueryEscape(text),
	)

	// const response = await fetch(apiUrl);
	resp, err := http.Get(apiUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// if (!response.ok) throw new Error(...)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API responded with status: %d", resp.StatusCode)
	}

	// const imageBuffer = await response.buffer();
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Upload image buffer to WhatsApp CDN
	uploadResp, err := client.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	cleanWho := who.ToNonAD()

	// Prepare WhatsApp Image Message
	imgMsg := &waE2E.ImageMessage{
		Caption:       proto.String(fmt.Sprintf("*@%s*", cleanWho.User)),
		Mimetype:      proto.String("image/png"),
		URL:           &uploadResp.URL,
		DirectPath:    &uploadResp.DirectPath,
		MediaKey:      uploadResp.MediaKey,
		FileEncSHA256: uploadResp.FileEncSHA256,
		FileSHA256:    uploadResp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(imageBytes))),
		ContextInfo: &waE2E.ContextInfo{
			MentionedJID: []string{cleanWho.String()},
		},
	}

	// await sock.sendMessage(...)
	_, err = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		ImageMessage: imgMsg,
	})
	return err
}