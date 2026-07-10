package main

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// Handler: .developer / .dev Command
func handleDeveloperCommand(client *whatsmeow.Client, v *events.Message) {
	// Send a crown emoji reaction
	react(client, v.Info.Chat, v.Info.ID, "👑")

	// Developer details (Apna number '923000000000' ki jagah enter karein)
	devNumber := "923195447147" 
	devLink := fmt.Sprintf("https://wa.me/%s", devNumber)

	devMessage := "👑 *𝐃𝐄𝐕𝐄𝐋𝐎𝐏𝐄𝐑 𝐈𝐍𝐅𝐎* 👑\n" +
		"━━━━━━━━━━━━━━━━━━━━\n" +
		"👤 *Name:* Bunny\n" +
		"📱 *WhatsApp:* " + devLink + "\n" +
		"💻 *Telegram:* t.me/bunny_devs\n" +
		"━━━━━━━━━━━━━━━━━━━━\n\n" +
		"😜 *Status:*\n" +
		"_Abhi tak to single ha, agar koi single larki ho to msg kero_ 😉💬\n\n" +
		"🚀 _Powered by: Bunny_"

	replyMessage(client, v, devMessage)
	
	// Success reaction
	react(client, v.Info.Chat, v.Info.ID, "✅")
}