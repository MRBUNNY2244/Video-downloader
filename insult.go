package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings" // Active use hai, isliye ab error nahi aayega
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func init() {
	// Random generator seed
	rand.Seed(time.Now().UnixNano())
}

// Helper: Safely extracts ContextInfo to identify replies or mentions
func getInsultContextInfo(msg *waE2E.Message) *waE2E.ContextInfo {
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

// Helper: Detect target user JID from replies, mentions, or manual phone number/name input
func getInsultTargetJID(v *events.Message, fullArgs string) types.JID {
	// 1. Agar text arguments ke andar manually "@bunny" likha ho (Case-insensitive)
	lowerArgs := strings.ToLower(fullArgs)
	if strings.Contains(lowerArgs, "bunny") {
		parsed, err := types.ParseJID("923195447147@s.whatsapp.net")
		if err == nil {
			return parsed
		}
	}

	ctxInfo := getInsultContextInfo(v.Message)
	if ctxInfo != nil {
		// 2. Quoted Message Sender
		if ctxInfo.Participant != nil {
			parsed, err := types.ParseJID(*ctxInfo.Participant)
			if err == nil {
				return parsed
			}
		}
		// 3. First Mentioned User JID
		if len(ctxInfo.MentionedJID) > 0 {
			parsed, err := types.ParseJID(ctxInfo.MentionedJID[0])
			if err == nil {
				return parsed
			}
		}
	}

	// 4. Extract manual number from text arguments (e.g. .insult 923195447147)
	cleanArg := ""
	for _, r := range fullArgs {
		if r >= '0' && r <= '9' {
			cleanArg += string(r)
		}
	}
	if len(cleanArg) >= 10 { // Minimal standard phone number length
		parsed, err := types.ParseJID(cleanArg + "@s.whatsapp.net")
		if err == nil {
			return parsed
		}
	}

	// Default: If no target found, default to sender
	return v.Info.Sender
}

// Handler: .insult Command
func handleInsultCommand(client *whatsmeow.Client, v *events.Message, fullArgs string) {
	// 1. Determine the target JID
	targetJID := getInsultTargetJID(v, fullArgs)
	cleanTarget := targetJID.ToNonAD()

	// 2. Owner Protection Shield (Locks number 923195447147 and name "Bunny" with BOLD text)
	lowerArgs := strings.ToLower(fullArgs)
	if cleanTarget.User == "923195447147" || strings.Contains(lowerArgs, "bunny") {
		// Bold output string: *BETA BAAP KI INSULT NAHI KERTA*
		replyMessage(client, v, "*BETA BAAP KI INSULT NAHI KERTA*")
		react(client, v.Info.Chat, v.Info.ID, "❌")
		return
	}

	react(client, v.Info.Chat, v.Info.ID, "😏")

	// 50 Sarcastic Roman Urdu Roasts
	insults := []string{
		"Aapko dekh kar lagta hai ke khuda ne dimaag bantte waqt aapko line mein khada hona bhula diya tha. 🧠❌",
		"Shakal achhi na ho toh insaan baat toh tameez se kar leta hai, par aapne toh dono mein compromise kiya hai. 🤐",
		"Jitna dimaag aapke paas hai, utne mein toh ek macchar bhi apna laptop nahi khol sakta. 🦟💻",
		"Aapki aqal aur dunya ki tarqi ka aapas mein door door tak koi talluq nahi hai. 📉🧠",
		"Suna hai dimaag lagane se barhta hai, par aapne toh dimaag ko tijori mein band karke rakha hua hai. 🔐💀",
		"Aapko dekh kar lagta hai ke biology ke evolution ne aapko chhor kar baaqi sab par kaam kiya hai. 🐵🚶‍♂️",
		"Kuch log paidaishi dheet hote hain, aur phir kuch log aap jaise hote hain jo isme PhD kar lete hain. 🎓😂",
		"Aapki baatein sun kar lagta hai ke headphone laga kar khamoshi sunna zyada behtar hai. 🎧🤐",
		"Agar stupidity ka koi award hota, toh aap hamesha bina kisi competition ke winner hote. 🏆🤡",
		"Apne dimaag ka dahi na karein, waise bhi dahi jamane ke liye dimaag ki zaroorat hoti hai. 🥛🧠",
		"Aapki tareef mein kya kahoon, aap toh bas dunya par ek bojh hain. 🌍🎒",
		"Suna hai har cheez ki koi na koi limit hoti hai, par aapki bewaqufi ki koi sarhad nahi. 🗺️🚫",
		"Aapki soch aur slow internet speed mein bilkul koi farq nahi hai, dono hi dimag kharab karte hain. 🐌🐢",
		"Aapko dekh kar lagta hai ke kash mein us waqt offline hota jab aap paida hue the. 🔌❌",
		"Aapki baatein sun kar dimaag ki dahi nahi, seedha lassi ban jati hai. 🥛🌀",
		"Agar khubsurati dimag se naapi jati, toh aap poori dunya ke sabse gareeb insaan hote. 💸💀",
		"Aapki khamoshi dunya ka sabse bada ehsaan hai, koshish karein ke yeh ehsaan hamesha rahe. 🤫",
		"Jitni speed se aap bewaqufi karte hain, utni speed se toh light bhi travel nahi karti. ⚡🚶‍♂️",
		"Aapka dimaag naya ka naya hai, lagta hai kabhi use hi nahi kiya. ✨🧠",
		"Aapki tareef karne ka dil toh bohot karta hai, par jhoot bolna meri fitrat mein nahi. 🤐🤷‍♂️",
		"Suna hai insaan apni ghaltiyon se seekhta hai, par aapne toh ghaltiyon ko hi apna rishon-dar bana liya hai. 👨‍👩‍👦❌",
		"Aapki aqal ghas charne gayi thi, aur lagta hai wahin ki ho kar reh gayi. 🌾🐐",
		"Aapki baatein sun kar dunya ke bade bade scientist bhi apni degrees phaarne par majboor ho jayein. 📄🎓",
		"Aapka dimaag toh bohot tez hai, par afsos wrong direction mein chal raha hai. 🏹🧭",
		"Aapki presence dunya mein bilkul wahi kaam karti hai jo phone mein 'Storage Full' ka notification karta hai. ⚠️📱",
		"Aap se behtar dimaag toh ek band pade calculator ke paas hota hai. 🔢💀",
		"Aapki aqal dekh kar lagta hai ke khuda ne aapko sirf body di hai, dimaag dena bhool gaye. 🧍‍♂️❌",
		"Aapki baatein kisi purane radio jaisi hain, jo bajta toh hai par samajh kuch nahi aata. 📻📡",
		"Aapko samjhana bilkul aisa hai jaise diwar par sir marna, nuksaan sirf apna hi hota hai. 🧱🤕",
		"Aapka dimaag toh 4G speed par chalta hai, par afsos aqal 2G par hi phasi hui hai. 📶🐢",
		"Aapki smile dekh kar lagta hai ke kash mask pehnne ka riwaj hamesha ke liye rehta. 😷🫣",
		"Aapki baaton mein na toh koi sar hota hai na paer, bas dimaag ka kachra hota hai. 🗑️🗣️",
		"Aapko dekh kar lagta hai ke dunya mein kuch log sirf space lene ke liye hi paida hote hain. 🌌🚀",
		"Aapki soch ka level bilkul wahi hai jo ground floor ke neeche basement ka hota hai. 📉🏠",
		"Suna hai hardwork se sab kuch mil jata hai, par aapne toh aqal ke mamle mein bilkul mehnat nahi ki. 🛠️🧠",
		"Aapki baatein sun kar dimaag ke cells khudkhushi karne lagte hain. 🧠🔫",
		"Aapko dekh kar lagta hai ke nature ne aap par kaam karte waqt copy-paste mein koi ghalti kar di hai. 💻❌",
		"Aapki tareef mein sirf yahi keh sakta hoon ke aap dunya ke sabse anokhe namune hain. 🏛️🏆",
		"Aapki bewaqufi dekh kar toh Einstein bhi apni qabar se uth kar aapko thapar marne aa jaye. ⚰️👋",
		"Aapka dimaag bilkul un-opened envelope jaisa hai, bilkul saaf aur mehfooz. ✉️🔒",
		"Aapki baatein sun kar lagta hai ke munh dho kar aana chahiye tha, shayad aqal thori saaf ho jati. 🧼🚿",
		"Aapko dekh kar lagta hai ke dunya mein abhi bhi 'Under Construction' signboard ki zaroorat hai. 🚧👷‍♂️",
		"Aapki aqal aur dunya ki haseen cheezon ka aapas mein koi rishta nahi hai. 🤷‍♂️💔",
		"Aapki baten kisi dawayi ki tarah hain, sunte hi neend aane lagti hai. 💊💤",
		"Aapka chehra dekh kar lagta hai ke kash camera mein 'auto-delete' ka option hota. 📸🗑️",
		"Aapko dekh kar lagta hai ke khuda ne aapko sirf entertainment ke liye hi banaya hai. 🎭🎪",
		"Aapki har baat mein ek ajeeb sa tanz hota hai, aur aapki har aqal mein ek ajeeb sa nuksaan. 📉",
		"Aap se baat karne se behtar hai ke insaan dhoop mein khada ho kar apni parchayi se baatein kare. ☀️👥",
		"Aapki bewaqufi dekh kar lagta hai ke dunya ab sach mein end hone wali hai. ☄️🌍",
		"Aapka dimaag bilkul empty space jaisa hai, jahan sirf khali hawa ghoomti rehti hai. 🌬️🌪️",
	}

	randomIndex := rand.Intn(len(insults))
	selectedInsult := insults[randomIndex]

	// Mentions highlights properly in the message using ExtendedTextMessage
	outputText := fmt.Sprintf("🔥 *YOUR IZZAT* 🔥\n━━━━━━━━━━━━━━━━━━━━\n\n@%s %s\n\n━━━━━━━━━━━━━━━━━━━━", cleanTarget.User, selectedInsult)

	_, err := client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(outputText),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: []string{cleanTarget.String()},
			},
		},
	})

	if err != nil {
		react(client, v.Info.Chat, v.Info.ID, "❌")
	} else {
		react(client, v.Info.Chat, v.Info.ID, "💀")
	}
}