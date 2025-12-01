package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"wago-backend/internal/model"
	"wago-backend/internal/webhook"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

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

		cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, phoneNumber, deviceInfo)

		// Notify WS
		cm.WSHub.SendToSession(sessionID, "status_update", map[string]interface{}{
			"status":       "connected",
			"phone_number": phoneNumber,
			"device_info":  deviceInfo,
		})

	case *events.Connected:
		// Update DB status if needed
		// cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, ...)
		// But we might not have phone number here if it's a reconnect.
		// Usually PairSuccess is for NEW login. Connected is for every connection.

		// Notify WS
		cm.WSHub.SendToSession(sessionID, "status_update", map[string]interface{}{
			"status": "connected",
		})

	case *events.LoggedOut:
		cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusDisconnected, "", nil)
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
		// Handle image caption
		if payload.Message == "" {
			payload.Message = v.Message.GetImageMessage().GetCaption()
			if v.Message.GetImageMessage() != nil {
				payload.MessageType = "image"
			}
		}

		// Filter out empty messages (e.g. status updates, protocol messages)
		if payload.Message == "" {
			return
		}

		// Send Webhook and Handle Response
		go func() {
			// Send Typing Indicator
			client := cm.GetClient(sessionID)
			if client != nil {
				// We need the JID of the sender (chat)
				chatJID := v.Info.Chat
				client.SendChatPresence(context.Background(), chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
			}

			response, err := cm.WebhookService.SendWebhook(session.WebhookURL, payload)

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
					}
				} else {
					fmt.Println("[Handler] Client is nil, cannot send response")
				}
			} else {
				fmt.Println("[Handler] Webhook response is empty, nothing to send.")
			}
		}()

		// Notify WS (optional, for debugging)
		msgBytes, _ := json.Marshal(v.Message)
		cm.WSHub.SendToSession(sessionID, "message_received", map[string]interface{}{
			"message": string(msgBytes),
		})
	}
}
