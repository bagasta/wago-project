package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wago-backend/internal/model"
	"wago-backend/internal/webhook"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

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

		// Handle extended text message (if conversation is empty)
		if payload.Message == "" {
			payload.Message = v.Message.GetExtendedTextMessage().GetText()
		}

		// Handle image message
		if imgMsg := v.Message.GetImageMessage(); imgMsg != nil {
			payload.MessageType = "image"
			if payload.Message == "" {
				payload.Message = imgMsg.GetCaption()
			}
		}

		// Filter out empty messages (e.g. status updates, protocol messages)
		if payload.Message == "" && payload.MessageType != "image" {
			return
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

		// Group Message Handling: Only respond if mentioned
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
			} else {
				fmt.Println("[GroupMsg] Client or Store ID is nil")
			}
		}

		// Send Webhook and Handle Response
		// Send Webhook and Handle Response
		go func(payload webhook.WebhookPayload) {
			// Check for image and download here
			if imgMsg := v.Message.GetImageMessage(); imgMsg != nil {
				fmt.Printf("[Handler] Found image message. Attempting to download...\n")
				client := cm.GetClient(sessionID)
				if client != nil {
					// Use timeout for download
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					data, err := client.Download(ctx, imgMsg)
					if err != nil {
						fmt.Printf("[Handler] Failed to download image: %v\n", err)
						payload.Message += fmt.Sprintf(" [Image Download Failed: %v]", err)
					} else {
						payload.MediaData = data
						payload.MediaMimeType = imgMsg.GetMimetype()
						// Determine extension from mimetype
						ext := "jpg" // default
						if strings.Contains(payload.MediaMimeType, "png") {
							ext = "png"
						} else if strings.Contains(payload.MediaMimeType, "jpeg") {
							ext = "jpg"
						} else if strings.Contains(payload.MediaMimeType, "webp") {
							ext = "webp"
						}
						payload.MediaName = fmt.Sprintf("image_%d.%s", v.Info.Timestamp.Unix(), ext)
						fmt.Printf("[Handler] Downloaded image successfully. Size: %d bytes, Mime: %s\n", len(data), payload.MediaMimeType)
					}
				} else {
					fmt.Printf("[Handler] Client is nil, cannot download image.\n")
					payload.Message += " [Image Download Failed: Client not found]"
				}
			}

			start := time.Now()
			// Send Typing Indicator
			client := cm.GetClient(sessionID)
			if client != nil {
				// We need the JID of the sender (chat)
				chatJID := v.Info.Chat
				client.SendChatPresence(context.Background(), chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
			}

			response, err := cm.WebhookService.SendWebhook(session.WebhookURL, payload)

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
					IsMention:           false, // We can refine this
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

			// Stop Typing Indicator
			if client != nil {
				chatJID := v.Info.Chat
				client.SendChatPresence(context.Background(), chatJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
			}

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
		}(payload)

		// Notify WS (optional, for debugging)
		msgBytes, _ := json.Marshal(v.Message)
		cm.WSHub.SendToSession(sessionID, "message_received", map[string]interface{}{
			"message": string(msgBytes),
		})
	}
}
