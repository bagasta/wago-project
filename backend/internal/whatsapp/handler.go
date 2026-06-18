package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"wago-backend/internal/model"
	"wago-backend/internal/webhook"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type inboundMedia struct {
	messageType string
	caption     string
	fileName    string
	mimeType    string
	download    whatsmeow.DownloadableMessage
}

// collectContextInfos gathers context info from common message types so we can check mentions in captions/text.
func collectContextInfos(msg *waProto.Message) []*waProto.ContextInfo {
	var contexts []*waProto.ContextInfo
	if msg.GetExtendedTextMessage() != nil {
		contexts = append(contexts, msg.GetExtendedTextMessage().GetContextInfo())
	}
	if msg.GetImageMessage() != nil {
		contexts = append(contexts, msg.GetImageMessage().GetContextInfo())
	}
	if msg.GetVideoMessage() != nil {
		contexts = append(contexts, msg.GetVideoMessage().GetContextInfo())
	}
	if msg.GetDocumentMessage() != nil {
		contexts = append(contexts, msg.GetDocumentMessage().GetContextInfo())
	}
	if msg.GetAudioMessage() != nil {
		contexts = append(contexts, msg.GetAudioMessage().GetContextInfo())
	}
	if msg.GetStickerMessage() != nil {
		contexts = append(contexts, msg.GetStickerMessage().GetContextInfo())
	}
	if msg.GetLocationMessage() != nil {
		contexts = append(contexts, msg.GetLocationMessage().GetContextInfo())
	}
	if msg.GetLiveLocationMessage() != nil {
		contexts = append(contexts, msg.GetLiveLocationMessage().GetContextInfo())
	}
	return contexts
}

// isMentioned checks both explicit mention lists and raw text for any of our JIDs (regular or LID).
func isMentioned(msg *waProto.Message, rawText string, targets []types.JID) bool {
	var searchTokens []string
	for _, jid := range targets {
		if jid.User == "" && jid.Server == "" {
			continue
		}
		// Base user
		searchTokens = append(searchTokens, jid.User)
		// Full JIDs
		searchTokens = append(searchTokens, jid.String())
		searchTokens = append(searchTokens, jid.ToNonAD().String())

		// Also include LID server form to catch mentions that use @lid even if our main JID is s.whatsapp.net
		if jid.Server != types.HiddenUserServer && jid.User != "" {
			lidJID := types.NewJID(jid.User, types.HiddenUserServer)
			searchTokens = append(searchTokens, lidJID.User, lidJID.String())
		}
	}

	// Check explicit mention lists in context infos.
	for _, ctx := range collectContextInfos(msg) {
		if ctx == nil {
			continue
		}
		for _, mentioned := range ctx.GetMentionedJID() {
			for _, t := range searchTokens {
				if strings.Contains(mentioned, t) {
					return true
				}
			}
		}
	}

	// Fallback: check plain text for @<number>
	text := strings.ToLower(rawText)
	for _, t := range searchTokens {
		if strings.Contains(text, "@"+strings.ToLower(t)) {
			return true
		}
	}
	return false
}

func mediaExtension(mimeType string, fallback string) string {
	mimeType = strings.ToLower(mimeType)
	switch {
	case strings.Contains(mimeType, "png"):
		return "png"
	case strings.Contains(mimeType, "jpeg"), strings.Contains(mimeType, "jpg"):
		return "jpg"
	case strings.Contains(mimeType, "webp"):
		return "webp"
	case strings.Contains(mimeType, "gif"):
		return "gif"
	case strings.Contains(mimeType, "pdf"):
		return "pdf"
	case strings.Contains(mimeType, "wordprocessingml"):
		return "docx"
	case strings.Contains(mimeType, "msword"):
		return "doc"
	case strings.Contains(mimeType, "spreadsheetml"):
		return "xlsx"
	case strings.Contains(mimeType, "excel"):
		return "xls"
	case strings.Contains(mimeType, "presentationml"):
		return "pptx"
	case strings.Contains(mimeType, "powerpoint"):
		return "ppt"
	case strings.Contains(mimeType, "mp4"):
		return "mp4"
	case strings.Contains(mimeType, "mpeg"):
		return "mp3"
	case strings.Contains(mimeType, "ogg"):
		return "ogg"
	case strings.Contains(mimeType, "webm"):
		return "webm"
	}
	return fallback
}

func defaultMediaExtension(prefix string) string {
	switch prefix {
	case "image":
		return "jpg"
	case "video":
		return "mp4"
	case "audio":
		return "ogg"
	case "sticker":
		return "webp"
	default:
		return "bin"
	}
}

func mediaFileName(prefix string, timestamp time.Time, mimeType string, providedName string) string {
	if providedName = strings.TrimSpace(providedName); providedName != "" {
		return filepath.Base(providedName)
	}
	return fmt.Sprintf("%s_%d.%s", prefix, timestamp.Unix(), mediaExtension(mimeType, defaultMediaExtension(prefix)))
}

func mediaLabel(messageType string) string {
	if messageType == "" {
		return "Media"
	}
	return strings.ToUpper(messageType[:1]) + messageType[1:]
}

func extractInboundMedia(msg *waProto.Message, timestamp time.Time) *inboundMedia {
	if msg == nil {
		return nil
	}
	if imgMsg := msg.GetImageMessage(); imgMsg != nil {
		return &inboundMedia{
			messageType: "image",
			caption:     imgMsg.GetCaption(),
			fileName:    mediaFileName("image", timestamp, imgMsg.GetMimetype(), ""),
			mimeType:    imgMsg.GetMimetype(),
			download:    imgMsg,
		}
	}
	if docMsg := msg.GetDocumentMessage(); docMsg != nil {
		return &inboundMedia{
			messageType: "document",
			caption:     docMsg.GetCaption(),
			fileName:    mediaFileName("document", timestamp, docMsg.GetMimetype(), docMsg.GetFileName()),
			mimeType:    docMsg.GetMimetype(),
			download:    docMsg,
		}
	}
	if videoMsg := msg.GetVideoMessage(); videoMsg != nil {
		return &inboundMedia{
			messageType: "video",
			caption:     videoMsg.GetCaption(),
			fileName:    mediaFileName("video", timestamp, videoMsg.GetMimetype(), ""),
			mimeType:    videoMsg.GetMimetype(),
			download:    videoMsg,
		}
	}
	if audioMsg := msg.GetAudioMessage(); audioMsg != nil {
		return &inboundMedia{
			messageType: "audio",
			fileName:    mediaFileName("audio", timestamp, audioMsg.GetMimetype(), ""),
			mimeType:    audioMsg.GetMimetype(),
			download:    audioMsg,
		}
	}
	if stickerMsg := msg.GetStickerMessage(); stickerMsg != nil {
		return &inboundMedia{
			messageType: "sticker",
			fileName:    mediaFileName("sticker", timestamp, stickerMsg.GetMimetype(), ""),
			mimeType:    stickerMsg.GetMimetype(),
			download:    stickerMsg,
		}
	}
	return nil
}

func isWebhookForwardableWithoutText(messageType string) bool {
	switch messageType {
	case "image", "document", "audio", "video", "sticker", "location", "live_location":
		return true
	default:
		return false
	}
}

func (cm *ClientManager) handleEvent(sessionID string, evt interface{}) {
	switch v := evt.(type) {
	case *events.PairSuccess:
		// Update DB
		jid := v.ID
		// Save FULL JID string (User@Server:DeviceID) to ensure we get the correct device later
		phoneNumber := jid.String()
		deviceInfo := &model.DeviceInfo{
			Platform:    v.Platform,
			DeviceModel: v.BusinessName, // Sometimes business name is here
		}

		fmt.Printf("PairSuccess: Saving session %s with JID %s\n", sessionID, phoneNumber)

		err := cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, &phoneNumber, deviceInfo)
		if err != nil {
			fmt.Printf("Failed to update session status: %v\n", err)
		} else {
			if updated, fetchErr := cm.SessionRepo.GetSessionByID(sessionID); fetchErr == nil && updated != nil {
				fmt.Printf("PairSuccess: session %s saved with phone_number=%s status=%s\n", sessionID, updated.PhoneNumber, updated.Status)
			}
		}

		// Notify WS
		cm.WSHub.SendToSession(sessionID, "status_update", map[string]interface{}{
			"status":       "connected",
			"phone_number": phoneNumber,
			"device_info":  deviceInfo,
		})

	case *events.Connected:
		// Ensure DB reflects connected status (covers reconnects where PairSuccess is not fired)
		var phoneNumber string
		// Try to get the JID from the in-memory client store
		client := cm.GetClient(sessionID)
		if client != nil && client.Store != nil && client.Store.ID != nil {
			phoneNumber = client.Store.ID.String()
		}

		// Fallback to existing DB value if we couldn't read from client
		if phoneNumber == "" {
			session, err := cm.SessionRepo.GetSessionByID(sessionID)
			if err == nil && session != nil {
				phoneNumber = session.PhoneNumber
			}
		}

		// Persist connected status + phone (if available)
		if err := cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, &phoneNumber, nil); err != nil {
			fmt.Printf("Failed to update session status on reconnect: %v\n", err)
		} else {
			if updated, fetchErr := cm.SessionRepo.GetSessionByID(sessionID); fetchErr == nil && updated != nil {
				fmt.Printf("Connected: session %s saved with phone_number=%s status=%s\n", sessionID, updated.PhoneNumber, updated.Status)
			}
		}

		// Notify WS
		cm.WSHub.SendToSession(sessionID, "status_update", map[string]interface{}{
			"status":       "connected",
			"phone_number": phoneNumber,
		})

	case *events.LoggedOut:
		empty := ""
		cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusDisconnected, &empty, nil)
		cm.WSHub.SendToSession(sessionID, "status_update", map[string]interface{}{
			"status": "disconnected",
		})

		// Remove from manager
		cm.mu.Lock()
		delete(cm.Clients, sessionID)
		cm.mu.Unlock()

	case *events.Message:
		// Handle incoming message
		fmt.Printf("Received message in session %s: %s\n", sessionID, v.Message.GetConversation())

		// Get Session to find Webhook URL
		session, err := cm.SessionRepo.GetSessionByID(sessionID)
		if err != nil {
			fmt.Printf("Error getting session for webhook: %v\n", err)
			return
		}

		// Construct Payload
		// Construct Payload
		payload := webhook.WebhookPayload{
			SessionID:   sessionID,
			From:        v.Info.Sender.User, // Phone number
			To:          "",                 // v.Info.Receiver is not available in MessageInfo. It's usually the connected user.
			Message:     v.Message.GetConversation(),
			Timestamp:   v.Info.Timestamp,
			IsGroup:     v.Info.IsGroup,
			PushName:    v.Info.PushName,
			MessageType: "text", // Simplify for now
		}

		// Attempt to resolve LID to Phone Number if Sender is a LID
		if v.Info.Sender.Server == types.HiddenUserServer {
			// If it's a LID, we want the User part (the UUID).
			// v.Info.Sender.User should be the UUID.
			// However, if the user says it changes, maybe they are seeing Device variations effectively?
			// Let's force using the Base ID for the field we send to avoid device suffix confusion.
			payload.SenderLID = v.Info.Sender.ToNonAD().User

			client := cm.GetClient(sessionID)
			if client != nil && client.Store != nil && client.Store.LIDs != nil {
				// Use Non-AD JID for lookup to ensure we match the User, not a specific device session
				pn, err := client.Store.LIDs.GetPNForLID(context.Background(), v.Info.Sender.ToNonAD())
				if err == nil && pn.User != "" {
					fmt.Printf("[Handler] Resolved LID %s (Device %d) to Phone %s\n", v.Info.Sender.User, v.Info.Sender.Device, pn.User)
					payload.From = pn.User
				} else {
					fmt.Printf("[Handler] Failed to resolve LID %s: %v\n", v.Info.Sender.User, err)
				}
			}
		} else if v.Info.Sender.Server == types.DefaultUserServer {
			// Sender is standard PN, try to resolve LID
			client := cm.GetClient(sessionID)
			if client != nil && client.Store != nil && client.Store.LIDs != nil {
				// Use Non-AD JID for lookup
				lid, err := client.Store.LIDs.GetLIDForPN(context.Background(), v.Info.Sender.ToNonAD())
				if err == nil && lid.User != "" {
					fmt.Printf("[Handler] Resolved Phone %s (Device %d) to LID %s\n", v.Info.Sender.User, v.Info.Sender.Device, lid.User)
					payload.SenderLID = lid.User
				} else {
					fmt.Printf("[Handler] No LID found for Phone %s: %v\n", v.Info.Sender.User, err)
				}
			}
		}

		// Handle extended text message (if conversation is empty)
		if payload.Message == "" {
			payload.Message = v.Message.GetExtendedTextMessage().GetText()
		}

		media := extractInboundMedia(v.Message, v.Info.Timestamp)
		if media != nil {
			payload.MessageType = media.messageType
			payload.MediaName = media.fileName
			payload.MediaMimeType = media.mimeType
			if payload.Message == "" {
				payload.Message = media.caption
			}
		}

		// Handle location message
		if locMsg := v.Message.GetLocationMessage(); locMsg != nil {
			payload.MessageType = "location"
			payload.Location = &webhook.LocationInfo{
				Latitude:  locMsg.GetDegreesLatitude(),
				Longitude: locMsg.GetDegreesLongitude(),
				Name:      locMsg.GetName(),
				URL:       locMsg.GetURL(),
				IsLive:    false,
			}
			if payload.Message == "" {
				payload.Message = fmt.Sprintf("📍 Location shared: %s (%.6f, %.6f)",
					locMsg.GetName(), locMsg.GetDegreesLatitude(), locMsg.GetDegreesLongitude())
			}
		}

		// Handle live location message
		if liveLocMsg := v.Message.GetLiveLocationMessage(); liveLocMsg != nil {
			payload.MessageType = "live_location"
			payload.Location = &webhook.LocationInfo{
				Latitude:  liveLocMsg.GetDegreesLatitude(),
				Longitude: liveLocMsg.GetDegreesLongitude(),
				IsLive:    true,
			}
			if payload.Message == "" {
				payload.Message = fmt.Sprintf("📍 Live location shared (%.6f, %.6f)",
					liveLocMsg.GetDegreesLatitude(), liveLocMsg.GetDegreesLongitude())
			}
		}

		// Filter out empty messages (e.g. status updates, protocol messages)
		if payload.Message == "" && !isWebhookForwardableWithoutText(payload.MessageType) {
			return
		}

		// Group Message Handling: Only respond if mentioned
		isMention := false
		if v.Info.IsGroup {
			if !session.IsGroupResponseEnabled {
				fmt.Printf("Ignoring group message from %s: group response disabled.\n", v.Info.Sender.User)
				return
			}

			client := cm.GetClient(sessionID)
			if client != nil && client.Store.ID != nil {
				targets := []types.JID{*client.Store.ID}
				if client.Store.LID.User != "" || client.Store.LID.Server != "" {
					targets = append(targets, client.Store.LID)
				}

				if !isMentioned(v.Message, payload.Message, targets) {
					fmt.Printf("Ignoring group message from %s: not mentioned. My JIDs: %v\n", v.Info.Sender.User, targets)
					return
				}
				isMention = true
			} else {
				fmt.Println("[GroupMsg] Client or Store ID is nil")
			}
		}

		// Log Message to DB
		go func() {
			msgLog := &model.MessageLog{
				SessionID:   sessionID,
				Direction:   "incoming",
				FromNumber:  payload.From,
				ToNumber:    "", // We don't have our own number easily accessible here without querying
				MessageType: payload.MessageType,
				Content:     payload.Message,
				IsGroup:     payload.IsGroup,
				Timestamp:   payload.Timestamp,
			}
			if payload.IsGroup {
				msgLog.GroupID = v.Info.Chat.User
				msgLog.GroupName = v.Info.PushName // Not accurate for group name, but PushName is sender name
			}
			if err := cm.AnalyticsRepo.LogMessage(msgLog); err != nil {
				fmt.Printf("Failed to log message: %v\n", err)
			}
		}()

		// Send Webhook and Handle Response
		// Send Webhook and Handle Response
		go func(payload webhook.WebhookPayload, media *inboundMedia) {
			if media != nil {
				fmt.Printf("[Handler] Found %s message. Attempting to download...\n", media.messageType)
				client := cm.GetClient(sessionID)
				if client != nil {
					// Use timeout for download
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					data, err := client.Download(ctx, media.download)
					if err != nil {
						fmt.Printf("[Handler] Failed to download %s: %v\n", media.messageType, err)
						payload.Message += fmt.Sprintf(" [%s Download Failed: %v]", mediaLabel(media.messageType), err)
					} else {
						payload.MediaData = data
						if payload.MediaMimeType == "" {
							payload.MediaMimeType = "application/octet-stream"
						}
						fmt.Printf("[Handler] Downloaded %s successfully. Size: %d bytes, Mime: %s\n", media.messageType, len(data), payload.MediaMimeType)
					}
				} else {
					fmt.Printf("[Handler] Client is nil, cannot download %s.\n", media.messageType)
					payload.Message += fmt.Sprintf(" [%s Download Failed: Client not found]", mediaLabel(media.messageType))
				}
			}

			start := time.Now()

			// Start looping typing indicator — refresh every 5s until webhook responds.
			// WhatsApp auto-clears "typing" after ~10–15s on most clients, so we re-send
			// periodically to keep it alive while the AI is thinking.
			client := cm.GetClient(sessionID)
			typingDone := make(chan struct{})
			if client != nil {
				chatJID := v.Info.Chat
				go func() {
					// Send immediately so there's no initial delay
					client.SendChatPresence(context.Background(), chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-typingDone:
							return
						case <-ticker.C:
							client.SendChatPresence(context.Background(), chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
						}
					}
				}()
			}

			response, err := cm.WebhookService.SendWebhook(session.WebhookURL, payload)

			// Stop typing loop as soon as webhook returns
			close(typingDone)
			if client != nil {
				chatJID := v.Info.Chat
				client.SendChatPresence(context.Background(), chatJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
			}

			// Calculate response time
			duration := time.Since(start).Milliseconds()

			// Log Analytics
			go func() {
				analytics := &model.Analytics{
					SessionID:           sessionID,
					MessageID:           v.Info.ID,
					FromNumber:          payload.From,
					MessageType:         payload.MessageType,
					IsGroup:             payload.IsGroup,
					IsMention:           isMention,
					WebhookSent:         true,
					WebhookSuccess:      err == nil,
					WebhookResponseTime: int(duration),
					WebhookStatusCode:   200, // Simplify for now, WebhookService should return status
				}
				if err != nil {
					analytics.ErrorMessage = err.Error()
					analytics.WebhookStatusCode = 500
				}
				if logErr := cm.AnalyticsRepo.LogAnalytics(analytics); logErr != nil {
					fmt.Printf("Failed to log analytics: %v\n", logErr)
				}
			}()

			if err != nil {
				fmt.Printf("Failed to send webhook: %v\n", err)
				return
			}

			// Send Response if available
			if response != "" {
				fmt.Printf("[Handler] Got response from webhook: %s\n", response)
				if client != nil {
					chatJID := v.Info.Chat
					fmt.Printf("[Handler] Sending message to %s\n", chatJID)

					// Send text message
					resp, err := client.SendMessage(context.Background(), chatJID, &waProto.Message{
						Conversation: proto.String(response),
					})
					if err != nil {
						fmt.Printf("[Handler] Failed to send response: %v\n", err)
					} else {
						fmt.Printf("[Handler] Response sent successfully. ID: %s\n", resp.ID)

						// Log Outgoing Message (AI Reply)
						go func() {
							msgLog := &model.MessageLog{
								SessionID:   sessionID,
								Direction:   "outgoing",
								FromNumber:  "", // It's us
								ToNumber:    chatJID.User,
								MessageType: "text",
								Content:     response,
								IsGroup:     v.Info.IsGroup,
								Timestamp:   time.Now(),
							}
							if v.Info.IsGroup {
								msgLog.GroupID = chatJID.User
								msgLog.GroupName = v.Info.PushName
							}
							if err := cm.AnalyticsRepo.LogMessage(msgLog); err != nil {
								fmt.Printf("Failed to log outgoing message: %v\n", err)
							}
						}()
					}
				} else {
					fmt.Println("[Handler] Client is nil, cannot send response")
				}
			} else {
				fmt.Println("[Handler] Webhook response is empty, nothing to send.")
			}
		}(payload, media)

		// Notify WS (optional, for debugging)
		msgBytes, _ := json.Marshal(v.Message)
		cm.WSHub.SendToSession(sessionID, "message_received", map[string]interface{}{
			"message": string(msgBytes),
		})
	}
}
