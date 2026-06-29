package main

import (
	"math/rand"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func init() {
	// Random generator ko seed karna taake har baar alag line aaye
	rand.Seed(time.Now().UnixNano())
}

// Handler: .flirt Command (Pure Love / Romantic Version)
func handleFlirtCommand(client *whatsmeow.Client, v *events.Message) {
	react(client, v.Info.Chat, v.Info.ID, "💖")

	// 55 Pure Romantic/Love Pickup Lines (No Tech/No Hacking)
	flirtLines := []string{
		"Aapki smile dekh kar dil ko ajeeb sa sukoon milta hai, jaise koi dua kabool ho gayi ho. ❤️",
		"Suna hai sabse khubsurat ehsaas mohabbat hai, aur mujhe yeh ehsaas aapko dekh kar hua. ✨",
		"Aapki aankhon mein aisi kashish hai ke insaan sab kuch bhool kar bas dekhta hi reh jaye. 👀",
		"Zindagi mein haseen log toh bohot mile, par dil ko sukoon sirf aapke paas aakar mila. 🌹",
		"Kya aap koi khwab hain? Kyun ke jab se aapko dekha hai, aankhein band karne ka dil hi nahi karta. 💭",
		"Aapki saadgi hi aapka sabse bada husn hai, kisi banawat ki aapko zaroorat hi nahi. 🌸",
		"Kuch log dil ko aise chhu lete hain ke phir unke bina har rasta adhoora lagta hai. 💖",
		"Suna hai chand par daag hai, par aapko dekh kar lagta hai chand ko jalan ho rahi hai. 🌙",
		"Zindagi haseen lagne lagti hai jab koi aap jaisa pyara sa humsafar khayalon mein aa jaye. 💞",
		"Aapki aawaz mein jo mithas hai, woh dil ke har dard ko sukoon de deti hai. 🌬️🎵",
		"Aapko dekh kar lagta hai jaise sardi ki dhoop mein koi thandi hawa ka jhonka aya ho. ☀️❄️",
		"Aapki aankhein sach bolti hain, aur mujhe unki har baat par yakeen ho jata hai. 👀✨",
		"Khuda ne shayad bohot fursat mein aapko banaya hai, har ada bemisaal hai. 🌸",
		"Aapki baatein sunte rehne ka dil karta hai, jaise koi pyari si ghazal chal rahi ho. 🎶📜",
		"Suna hai dunya mein 7 ajoobe hain, par lagta hai unho ne abhi tak aapko nahi dekha. 🌍🏆",
		"Zindagi mein sab kuch thik ho jata hai, jab aapki ek pyari si smile samne aa jaye. 😊💖",
		"Aapki saadgi mere dil ko touch kar jati hai, sach mein aap bohot pyari hain. 🥺❤️",
		"Aapke sath guzra har lamha mere liye sabse bada tohfa hai. 🎁💞",
		"Mujhe dunya ki kisi cheez ki chahat nahi, bas aapka sath hamesha ke liye chahiye. 🤝💑",
		"Aapki hansi se mera poora din haseen ban jata hai. 🌟",
		"Suna hai saccha pyar naseeb walon ko milta hai, kya main khud ko naseeb wala samajh loon? 😉💗",
		"Aapki aankhon mein jo chamak hai, woh aasmaan ke taron mein bhi nahi milti. ✨👀",
		"Aapko dekh kar dil bas yahi kehta hai ke kash waqt yahin ruk jaye. ⏳❤️",
		"Mohabbat kya hoti hai mujhe nahi pata tha, par jab se aapko dekha dil dhadakna seekh gaya. 💓",
		"Aapki har baat mere dil ki gehraiyoon mein utar jati hai. 🎯💞",
		"Aap koi phool hain kya? Kyun ke jab se aap aaye hain, zindagi mehsus hone lagi hai. 🌸🍃",
		"Aapki aankhon mein khona toh aasan hai, par unse bahar nikalna namumkin hai. 👀🫠",
		"Zindagi mein hazaron khwahishein theen, par ab bas aap par hi aakar ruk gayeen hain. ❤️🔐",
		"Aapki baten sun kar dil ko lagta hai jaise koi dua qabool ho gayi ho. 🤲✨",
		"Suna hai barish mein mitti ki khushbu pyari hoti hai, par aapki khushbu sabse alag hai. 🌧️🌸",
		"Aapki har ek ada par dil fida hone ko taiyar rehta hai. 😍",
		"Aapki dosti mere liye sabse pyara ehsaas hai, aur aapka sath sabse bari khushi. 💖",
		"Aapko paana toh ek khwab lagta hai, par khwab dekhna humne chhoda nahi. 💭💫",
		"Aapki aankhon mein ek ajeeb sa sukoon hai, jo dunya ki kisi corner mein nahi mila. 👀🛋️",
		"Aapki smile dunya ki sabse khubsurat tasveer hai. 🖼️😊",
		"Mere dil ka har ek hissa bas aapke naam se hi dhadakta hai. 🫀💞",
		"Aapki baten kisi thandi chaoon ki tarah hain jo thakan door kar deti hain. 🌳💨",
		"Aapko dekh kar lagta hai ke dunya sach mein bohot haseen hai. 🌍🌸",
		"Aapke bina har ek pal adhoora sa lagta hai, jaise koi rasta bina manzil ke ho. 🗺️👣",
		"Aapki saadgi mere dil mein aise bas gayi hai ke ab nikalna mushkil hai. 👉❤️",
		"Suna hai haseen log bohot ziddi hote hain, par aap toh bilkul masoom hain. 🥺✨",
		"Aapki smile mere dil ke saare gham mita deti hai. 😊🛡️",
		"Aapki khushbu se lagta hai ke bahar ka mausam har waqt rehta hai. 🌸🍂",
		"Aapko khone ka darr hi shayad mujhe batata hai ke main aap se kitni mohabbat karta hoon. 💔💗",
		"Aapki aankhein jaise koi gehri jheel hon, jisme doobna haseen lagta hai. 🌊👀",
		"Aapki baten mere dil ki har ek dharakan ko sukoon deti hain. 🎶💓",
		"Mera har ek khwab aapki hi galiyon se guzarta hai. 🏘️💭",
		"Aap dunya ki baki ladkiyon se bohot alag aur bohot pyari hain. ✨🥇",
		"Aapki hansi se dunya ka andhera bhi roshan ho jata hai. 💡😍",
		"Aapki chahat meri zindagi ka sabse khubsurat hissa ban chuki hai. ❤️",
		"Suna hai sachi mohabbat bar bar nahi hoti, par mujhe lagta hai mujhe aap se har roz hoti hai. 💖✨",
		"Aapki smile mere pure din ki thakan mita deti hai. 🥰💤",
		"Aapki aawaz sun kar mera dil aise jhoomta hai jaise barish mein peacock. 🦚🌧️",
		"Mera har din tab haseen banta hai jab aapki ek jhalak dekhne ko mile. 🌅✨",
		"Agar main poori zindagi bhi aapki tareef karoon, toh bhi kam hai. ✍️📄",
	}

	// List me se ek random line select karna
	randomIndex := rand.Intn(len(flirtLines))
	selectedLine := flirtLines[randomIndex]

	replyMessage(client, v, selectedLine)
	react(client, v.Info.Chat, v.Info.ID, "🥰")
}