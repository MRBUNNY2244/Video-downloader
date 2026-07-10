package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// ─────────────────────────────────────────
// MUTE / UNMUTE
// .mute  → sirf admins likh sakein (announce mode on)
// .unmute → sab likh sakein
// ─────────────────────────────────────────
func handleMute(client *whatsmeow.Client, v *events.Message) {
	err := client.SetGroupAnnounce(context.Background(), v.Info.Chat, true)
	if err != nil {
		replyMessage(client, v, "❌ Failed! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔇")
	replyMessage(client, v, "🔇 *Group Muted!*\nOnly Admins can send messages now.")
}

func handleUnmute(client *whatsmeow.Client, v *events.Message) {
	err := client.SetGroupAnnounce(context.Background(), v.Info.Chat, false)
	if err != nil {
		replyMessage(client, v, "❌ Failed! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔊")
	replyMessage(client, v, "🔊 *Group Unmuted!*\nAll members can send messages now.")
}

// ─────────────────────────────────────────
// LOCK / UNLOCK  (group info editing lock)
// .lock  → sirf admins group info edit kar sakein
// .unlock → sab members edit kar sakein
// ─────────────────────────────────────────
func handleLock(client *whatsmeow.Client, v *events.Message) {
	err := client.SetGroupLocked(context.Background(), v.Info.Chat, true)
	if err != nil {
		replyMessage(client, v, "❌ Failed! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔒")
	replyMessage(client, v, "🔒 *Group Locked!*\nOnly Admins can edit group info now.")
}

func handleUnlock(client *whatsmeow.Client, v *events.Message) {
	err := client.SetGroupLocked(context.Background(), v.Info.Chat, false)
	if err != nil {
		replyMessage(client, v, "❌ Failed! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔓")
	replyMessage(client, v, "🔓 *Group Unlocked!*\nAll members can edit group info now.")
}

// ─────────────────────────────────────────
// KICKALL
// Sab non-admin members ko kick karta hai
// ─────────────────────────────────────────
func handleKickAll(client *whatsmeow.Client, v *events.Message) {
	groupInfo, err := client.GetGroupInfo(context.Background(), v.Info.Chat)
	if err != nil {
		replyMessage(client, v, "❌ Could not fetch group info.")
		return
	}

	botJID := client.Store.ID.ToNonAD()
	var toKick []types.JID
	for _, p := range groupInfo.Participants {
		if p.IsAdmin || p.IsSuperAdmin {
			continue
		}
		if p.JID.User == botJID.User {
			continue
		}
		toKick = append(toKick, p.JID)
	}

	if len(toKick) == 0 {
		replyMessage(client, v, "ℹ️ No non-admin members to kick.")
		return
	}

	replyMessage(client, v, fmt.Sprintf("⚠️ Kicking *%d* members... Please wait.", len(toKick)))

	kicked := 0
	for _, jid := range toKick {
		_, err := client.UpdateGroupParticipants(context.Background(), v.Info.Chat, []types.JID{jid}, whatsmeow.ParticipantChangeRemove)
		if err == nil {
			kicked++
		}
		time.Sleep(500 * time.Millisecond) // flood se bachao
	}

	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *Kickall Done!*\n%d/%d members kicked.", kicked, len(toKick)))
}

// ─────────────────────────────────────────
// GROUPDESC
// .groupdesc <new description>
// ─────────────────────────────────────────
func handleGroupDesc(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ Usage: `.groupdesc <new description>`")
		return
	}
	err := client.SetGroupDescription(context.Background(), v.Info.Chat, args)
	if err != nil {
		replyMessage(client, v, "❌ Failed! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, "✅ *Group description updated!*")
}

// ─────────────────────────────────────────
// GROUPLINK
// .grouplink       → invite link show karo
// .grouplink reset → link reset karo
// ─────────────────────────────────────────
func handleGroupLink(client *whatsmeow.Client, v *events.Message, args string) {
	if strings.ToLower(strings.TrimSpace(args)) == "reset" {
		_, err := client.RevokeGroupInvite(context.Background(), v.Info.Chat)
		if err != nil {
			replyMessage(client, v, "❌ Failed to reset link! Bot must be an Admin.")
			return
		}
		react(client, v.Info.Chat, v.Info.ID, "🔄")
		replyMessage(client, v, "🔄 *Group invite link has been reset!*")
		return
	}

	code, err := client.GetGroupInviteLink(context.Background(), v.Info.Chat, false)
	if err != nil {
		replyMessage(client, v, "❌ Failed to get link! Bot must be an Admin.")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "🔗")
	replyMessage(client, v, fmt.Sprintf("🔗 *Group Invite Link:*\nhttps://chat.whatsapp.com/%s\n\n_Type `.grouplink reset` to generate a new link._", code))
}

// ─────────────────────────────────────────
// GROUPINFO
// Group ki full info show karta hai
// ─────────────────────────────────────────
func handleGroupInfo(client *whatsmeow.Client, v *events.Message) {
	groupInfo, err := client.GetGroupInfo(context.Background(), v.Info.Chat)
	if err != nil {
		replyMessage(client, v, "❌ Could not fetch group info.")
		return
	}

	admins := 0
	for _, p := range groupInfo.Participants {
		if p.IsAdmin || p.IsSuperAdmin {
			admins++
		}
	}

	created := groupInfo.GroupCreated.Format("02 Jan 2006")
	desc := groupInfo.Topic
	if desc == "" {
		desc = "_No description set_"
	}

	text := fmt.Sprintf(`┏━━〔 📋 *GROUP INFO* 〕━━┈
┃ 📛 *Name:* %s
┃ 👥 *Members:* %d
┃ 👑 *Admins:* %d
┃ 📅 *Created:* %s
┃ 🔒 *Locked:* %v
┃ 📢 *Announce:* %v
┃
┃ 📝 *Description:*
┃ %s
┗━━━━━━━━━━━━━━━━━━━┈`,
		groupInfo.Name,
		len(groupInfo.Participants),
		admins,
		created,
		groupInfo.IsLocked,
		groupInfo.IsAnnounce,
		strings.ReplaceAll(desc, "\n", "\n┃ "),
	)

	react(client, v.Info.Chat, v.Info.ID, "📋")
	replyMessage(client, v, text)
}

// ─────────────────────────────────────────
// GETPIC
// .getpic → reply ya tag karo, profile pic download hogi
// ─────────────────────────────────────────
func handleGetPic(client *whatsmeow.Client, v *events.Message, args string) {
	targetJID, ok := getTargetJID(v, args)
	if !ok {
		// Agar koi target nahi, sender ki pic lo
		targetJID = v.Info.Sender.ToNonAD()
	}

	picInfo, err := client.GetProfilePictureInfo(context.Background(), targetJID, &whatsmeow.GetProfilePictureParams{Preview: false})
	if err != nil || picInfo == nil {
		replyMessage(client, v, "❌ No profile picture found or it's private.")
		return
	}

	react(client, v.Info.Chat, v.Info.ID, "🖼️")
	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:      proto.String(picInfo.URL),
			Mimetype: proto.String("image/jpeg"),
			Caption:  proto.String(fmt.Sprintf("📸 Profile picture of @%s", targetJID.User)),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: []string{targetJID.String()},
			},
		},
	})
}

// ─────────────────────────────────────────
// TAGADMIN
// Sirf admins ko tag karta hai
// ─────────────────────────────────────────
func handleTagAdmin(client *whatsmeow.Client, v *events.Message, args string) {
	groupInfo, err := client.GetGroupInfo(context.Background(), v.Info.Chat)
	if err != nil {
		replyMessage(client, v, "❌ Could not fetch group info.")
		return
	}

	var mentions []string
	var sb strings.Builder
	sb.WriteString("📢 *TAGGING ADMINS*\n\n")
	if args != "" {
		sb.WriteString(fmt.Sprintf("💬 *Message:* %s\n\n", args))
	}

	for _, p := range groupInfo.Participants {
		if p.IsAdmin || p.IsSuperAdmin {
			mentions = append(mentions, p.JID.String())
			sb.WriteString(fmt.Sprintf("👑 @%s\n", p.JID.User))
		}
	}

	if len(mentions) == 0 {
		replyMessage(client, v, "ℹ️ No admins found in this group.")
		return
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
// POLL
// .poll Question | Option1 | Option2 | Option3
// ─────────────────────────────────────────
func handlePoll(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:*\n`.poll Question | Option1 | Option2 | Option3`\n\nExample:\n`.poll Favourite color? | Red | Blue | Green`")
		return
	}

	parts := strings.Split(args, "|")
	if len(parts) < 3 {
		replyMessage(client, v, "❌ At least *1 question* and *2 options* required!\n\nExample:\n`.poll Favourite color? | Red | Blue`")
		return
	}

	question := strings.TrimSpace(parts[0])
	var options []*waProto.PollCreationMessage_Option
	for _, opt := range parts[1:] {
		trimmed := strings.TrimSpace(opt)
		if trimmed != "" {
			options = append(options, &waProto.PollCreationMessage_Option{
				OptionName: proto.String(trimmed),
			})
		}
	}

	if len(options) < 2 {
		replyMessage(client, v, "❌ At least *2 valid options* required!")
		return
	}

	react(client, v.Info.Chat, v.Info.ID, "📊")
	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		PollCreationMessage: &waProto.PollCreationMessage{
			Name:                proto.String(question),
			Options:             options,
			SelectableOptionsCount: proto.Uint32(1),
		},
	})
}

// ─────────────────────────────────────────
// SETWELCOME / SETGOODBYE
// Custom welcome/goodbye messages
// ─────────────────────────────────────────

var customWelcome = make(map[string]string)
var customGoodbye = make(map[string]string)
var msgMu sync.Mutex

func handleSetWelcome(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.setwelcome Welcome @user to @group! 🎉`\n\nVariables:\n`@user` → Member name\n`@group` → Group name")
		return
	}
	msgMu.Lock()
	customWelcome[v.Info.Chat.String()] = args
	msgMu.Unlock()
	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *Welcome message set!*\n\nPreview:\n_%s_", args))
}

func handleSetGoodbye(client *whatsmeow.Client, v *events.Message, args string) {
	if args == "" {
		replyMessage(client, v, "❌ *Usage:* `.setgoodbye Goodbye @user! 👋`\n\nVariables:\n`@user` → Member name\n`@group` → Group name")
		return
	}
	msgMu.Lock()
	customGoodbye[v.Info.Chat.String()] = args
	msgMu.Unlock()
	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *Goodbye message set!*\n\nPreview:\n_%s_", args))
}

// GetCustomWelcome/Goodbye — group.go ke welcome handler se call hoga
func GetCustomWelcome(groupJID string) string {
	msgMu.Lock()
	defer msgMu.Unlock()
	return customWelcome[groupJID]
}

func GetCustomGoodbye(groupJID string) string {
	msgMu.Lock()
	defer msgMu.Unlock()
	return customGoodbye[groupJID]
}

// ─────────────────────────────────────────
// JOIN REQUESTS: requests / accept / reject
// ─────────────────────────────────────────
func handleRequests(client *whatsmeow.Client, v *events.Message) {
	reqs, err := client.GetGroupRequestParticipants(context.Background(), v.Info.Chat)
	if err != nil {
		replyMessage(client, v, "❌ Could not fetch join requests. Bot must be Admin.")
		return
	}
	if len(reqs) == 0 {
		replyMessage(client, v, "ℹ️ No pending join requests.")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 *Pending Join Requests:* %d\n\n", len(reqs)))
	for i, jid := range reqs {
		sb.WriteString(fmt.Sprintf("%d. @%s\n", i+1, jid.User))
	}
	sb.WriteString("\n_Use `.accept all` or `.reject all` to manage._")

	var mentions []string
	for _, jid := range reqs {
		mentions = append(mentions, jid.String())
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

func handleAccept(client *whatsmeow.Client, v *events.Message, args string) {
	reqs, err := client.GetGroupRequestParticipants(context.Background(), v.Info.Chat)
	if err != nil || len(reqs) == 0 {
		replyMessage(client, v, "❌ No pending requests or bot is not Admin.")
		return
	}

	var toAccept []types.JID
	if strings.ToLower(args) == "all" || args == "" {
		toAccept = reqs
	} else {
		targetJID, ok := getTargetJID(v, args)
		if !ok {
			replyMessage(client, v, "❌ Usage: `.accept all` or reply/tag someone.")
			return
		}
		toAccept = []types.JID{targetJID}
	}

	_, err = client.UpdateGroupParticipants(context.Background(), v.Info.Chat, toAccept, whatsmeow.ParticipantChangeAdd)
	if err != nil {
		replyMessage(client, v, "❌ Failed to accept requests!")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *%d join request(s) accepted!*", len(toAccept)))
}

func handleReject(client *whatsmeow.Client, v *events.Message, args string) {
	reqs, err := client.GetGroupRequestParticipants(context.Background(), v.Info.Chat)
	if err != nil || len(reqs) == 0 {
		replyMessage(client, v, "❌ No pending requests or bot is not Admin.")
		return
	}

	var toReject []types.JID
	if strings.ToLower(args) == "all" || args == "" {
		toReject = reqs
	} else {
		targetJID, ok := getTargetJID(v, args)
		if !ok {
			replyMessage(client, v, "❌ Usage: `.reject all` or reply/tag someone.")
			return
		}
		toReject = []types.JID{targetJID}
	}

	_, err = client.UpdateGroupParticipants(context.Background(), v.Info.Chat, toReject, whatsmeow.ParticipantChangeRemove)
	if err != nil {
		replyMessage(client, v, "❌ Failed to reject requests!")
		return
	}
	react(client, v.Info.Chat, v.Info.ID, "✅")
	replyMessage(client, v, fmt.Sprintf("✅ *%d join request(s) rejected!*", len(toReject)))
}

// ─────────────────────────────────────────
// OPENTIME / CLOSETIME
// Schedule group open/close at a specific time
// Format: .opentime 08:30  (24hr)
// ─────────────────────────────────────────

type scheduleEntry struct {
	cancel chan struct{}
}

var schedules = make(map[string]*scheduleEntry)
var scheduleMu sync.Mutex

func handleOpenTime(client *whatsmeow.Client, v *events.Message, args string) {
	scheduleGroupAt(client, v, args, false)
}

func handleCloseTime(client *whatsmeow.Client, v *events.Message, args string) {
	scheduleGroupAt(client, v, args, true)
}

func scheduleGroupAt(client *whatsmeow.Client, v *events.Message, args string, mute bool) {
	action := "open"
	emoji := "🔊"
	if mute {
		action = "close"
		emoji = "🔇"
	}

	if args == "" {
		replyMessage(client, v, fmt.Sprintf("❌ *Usage:* `.%stime HH:MM`\nExample: `.%stime 08:30`", action, action))
		return
	}

	var hour, minute int
	_, err := fmt.Sscanf(strings.TrimSpace(args), "%d:%d", &hour, &minute)
	if err != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		replyMessage(client, v, "❌ Invalid time! Use 24-hour format. Example: `08:30` or `20:00`")
		return
	}

	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if target.Before(now) {
		target = target.Add(24 * time.Hour) // kal ke liye schedule
	}
	dur := time.Until(target)

	// Purana schedule cancel karo agar tha
	key := fmt.Sprintf("%s_%s", v.Info.Chat.String(), action)
	scheduleMu.Lock()
	if old, exists := schedules[key]; exists {
		close(old.cancel)
	}
	entry := &scheduleEntry{cancel: make(chan struct{})}
	schedules[key] = entry
	scheduleMu.Unlock()

	react(client, v.Info.Chat, v.Info.ID, emoji)
	replyMessage(client, v, fmt.Sprintf("%s *Scheduled!*\nGroup will *%s* at `%02d:%02d` (in %s)", emoji, action, hour, minute, dur.Round(time.Minute)))

	go func() {
		select {
		case <-time.After(dur):
			if mute {
				client.SetGroupAnnounce(context.Background(), v.Info.Chat, true)
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					Conversation: proto.String("🔇 *Group closed automatically by Bunny MD.*"),
				})
			} else {
				client.SetGroupAnnounce(context.Background(), v.Info.Chat, false)
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					Conversation: proto.String("🔊 *Group opened automatically by Bunny MD.*"),
				})
			}
		case <-entry.cancel:
			// cancelled
		}
	}()
}
