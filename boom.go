package main

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto" // 'go get google.golang.org/protobuf/proto' run karein agar error aaye
)

// Handler: .boom Command (Dynamic Message Editing)
func handleBoomCommand(client *whatsmeow.Client, v *events.Message) {
	react(client, v.Info.Chat, v.Info.ID, "💣")

	// 1. Pehla message send karein (Jo '0' se start hoga)
	resp, err := client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		Conversation: proto.String("0"),
	})
	if err != nil {
		react(client, v.Info.Chat, v.Info.ID, "❌")
		return
	}

	// 2. 1 se lekar 5 tak counting edit sequence
	for i := 1; i <= 5; i++ {
		time.Sleep(1 * time.Second)
		
		text := fmt.Sprintf("%d", i)
		
		// Message ko edit karne ke liye BuildEdit helper ka use
		editMsg := client.BuildEdit(v.Info.Chat, resp.ID, &waE2E.Message{
			Conversation: proto.String(text),
		})
		_, _ = client.SendMessage(context.Background(), v.Info.Chat, editMsg)
	}

	// 3. 5 seconds poore hone par final update
	time.Sleep(1 * time.Second)
	finalText := "💥 *BOOM!*\n\nkeisa LGA Mera mjak 😜"
	
	finalEdit := client.BuildEdit(v.Info.Chat, resp.ID, &waE2E.Message{
		Conversation: proto.String(finalText),
	})
	_, _ = client.SendMessage(context.Background(), v.Info.Chat, finalEdit)
	
	react(client, v.Info.Chat, v.Info.ID, "💥")
}