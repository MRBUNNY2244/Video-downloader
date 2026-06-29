package main

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// Handler: .larkiptana Command
func handleLarkiPtanaCommand(client *whatsmeow.Client, v *events.Message) {
	// Funny clown emoji reaction
	react(client, v.Info.Chat, v.Info.ID, "🤡")

	// Fixed exact text requested in bold with expressing emojis
	replyMessage(client, v, "*ABA SHAKAL DEKHI HA MUU DHO KAR AA, SHAKAL KYA HAI AND LARKIYAN CHAYIE INKO!* 🤡👣")

	// Mark as processed with a cross reaction
	react(client, v.Info.Chat, v.Info.ID, "❌")
}