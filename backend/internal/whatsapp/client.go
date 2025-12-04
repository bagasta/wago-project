package whatsapp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"wago-backend/internal/config"
	"wago-backend/internal/model"
	"wago-backend/internal/repository"
	"wago-backend/internal/webhook"
	"wago-backend/internal/websocket"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientManager struct {
	Clients        map[string]*whatsmeow.Client
	Config         *config.Config
	SessionRepo    *repository.SessionRepository
	AnalyticsRepo  *repository.AnalyticsRepository
	WSHub          *websocket.Hub
	WebhookService *webhook.WebhookService
	Container      *sqlstore.Container
	mu             sync.RWMutex
}

func NewClientManager(cfg *config.Config, sessionRepo *repository.SessionRepository, analyticsRepo *repository.AnalyticsRepository, wsHub *websocket.Hub, webhookService *webhook.WebhookService) *ClientManager {
	// Initialize whatsmeow SQL store
	dbLog := waLog.Stdout("Database", cfg.LogLevel, true)
	container, err := sqlstore.New(context.Background(), "postgres", cfg.DatabaseURL, dbLog)
	if err != nil {
		panic(err)
	}

	return &ClientManager{
		Clients:        make(map[string]*whatsmeow.Client),
		Config:         cfg,
		SessionRepo:    sessionRepo,
		AnalyticsRepo:  analyticsRepo,
		WSHub:          wsHub,
		WebhookService: webhookService,
		Container:      container,
	}
}

// normalizeSessionJID tries to turn whatever is stored in the DB into a valid JID that includes server (and device if present).
// types.ParseJID doesn't error on plain numbers, so we additionally ensure the user part is present.
func normalizeSessionJID(raw string) (types.JID, error) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return types.JID{}, fmt.Errorf("empty JID string")
	}

	jid, err := types.ParseJID(cleaned)
	if err == nil && jid.User != "" {
		// Ensure default server is set if somehow missing.
		if jid.Server == "" {
			jid.Server = types.DefaultUserServer
		}
		return jid, nil
	}

	// Fallback for bare phone numbers or other invalid formats: assume default WA server.
	if !strings.Contains(cleaned, "@") {
		cleaned = cleaned + "@" + types.DefaultUserServer
	}

	jid, err = types.ParseJID(cleaned)
	if err != nil {
		return types.JID{}, err
	}
	if jid.User == "" {
		return types.JID{}, fmt.Errorf("failed to parse user part from JID: %s", raw)
	}
	return jid, nil
}

func (cm *ClientManager) GetClient(sessionID string) *whatsmeow.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.Clients[sessionID]
}

func (cm *ClientManager) Connect(sessionID string) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.Clients[sessionID]; ok {
		if client.IsConnected() {
			return "connected", nil
		}
		// If client exists but not connected, we might want to reconnect?
		// But for now, let's assume if it's in the map, it's being handled.
		// However, if it's in the map but disconnected, we should probably let it proceed to reconnect logic?
		// But NewClient creates a new instance.
		// Let's just return "connected" if it's in the map for simplicity, or "connecting".
		return "connected", nil
	}

	// Get device store
	// ... (rest of logic remains same until return)

	var deviceStore *store.Device
	session, err := cm.SessionRepo.GetSessionByID(sessionID)
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", fmt.Errorf("session not found")
	}

	ctx := context.Background()

	if session.PhoneNumber != "" {
		jid, err := normalizeSessionJID(session.PhoneNumber)
		if err != nil {
			fmt.Printf("Invalid stored JID for session %s (%s): %v\n", sessionID, session.PhoneNumber, err)
		} else {
			deviceStore, err = cm.Container.GetDevice(ctx, jid)
			if err != nil {
				fmt.Printf("Device lookup failed for %s: %v\n", jid.String(), err)
			}

			// If direct lookup failed (e.g. stored JID missing device ID), search by user/server.
			if deviceStore == nil {
				devices, listErr := cm.Container.GetAllDevices(ctx)
				if listErr != nil {
					fmt.Printf("Failed to list devices for session %s: %v\n", sessionID, listErr)
				} else {
					for _, dev := range devices {
						if dev.ID.User == jid.User && dev.ID.Server == jid.Server {
							deviceStore = dev
							// Persist the full JID (with device) so next reconnect uses the exact match.
							if dev.ID.String() != session.PhoneNumber {
								if dev.ID.String() != session.PhoneNumber {
									ph := dev.ID.String()
									cm.SessionRepo.UpdateSessionStatus(sessionID, session.Status, &ph, session.DeviceInfo)
								}
							}
							break
						}
					}
				}
			}
		}
	}

	if deviceStore == nil {
		// New device (QR mode)
		deviceStore = cm.Container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", cm.Config.LogLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Add event handler
	client.AddEventHandler(func(evt interface{}) {
		cm.handleEvent(sessionID, evt)
	})

	cm.Clients[sessionID] = client

	// Connect
	if client.Store.ID == nil {
		// No ID means not logged in.
		// Get QR Channel
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return "", err
		}

		// Listen for QR
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					// Send QR to WebSocket
					cm.WSHub.SendToSession(sessionID, "qr_update", map[string]interface{}{
						"qr_code":    evt.Code,
						"expires_in": 60, // approximate
					})

					// Update DB status to 'qr'
					cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusQR, nil, nil)
				} else {
					// Timeout or success?
					// Success is handled by EventHandler
				}
			}
		}()
		return "qr", nil
	} else {
		// Already logged in
		err = client.Connect()
		if err != nil {
			return "", err
		}
		// Update status just in case
		// cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, client.Store.ID.User, nil)
		return "connected", nil
	}
}

func (cm *ClientManager) disconnect(sessionID string, updateStatus bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.Clients[sessionID]; ok {
		client.Disconnect()
		delete(cm.Clients, sessionID)
		if updateStatus {
			cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusDisconnected, nil, nil)
		}
	}
}

// Disconnect is used for user-triggered session stop; it updates DB status.
func (cm *ClientManager) Disconnect(sessionID string) {
	cm.disconnect(sessionID, true)
}

// Shutdown disconnects all active clients gracefully.
func (cm *ClientManager) Shutdown() {
	cm.mu.RLock()
	ids := make([]string, 0, len(cm.Clients))
	for id := range cm.Clients {
		ids = append(ids, id)
	}
	cm.mu.RUnlock()

	for _, id := range ids {
		// Do not overwrite status/phone_number during shutdown so auto-reconnect still works
		cm.disconnect(id, false)
	}
}

// ReconnectAllSessions reconnects all sessions that are marked as connected in the DB
func (cm *ClientManager) ReconnectAllSessions() {
	// Try reconnecting any session that has a stored JID (phone_number),
	// even if status wasn't left as "connected" due to an unclean shutdown.
	sessions, err := cm.SessionRepo.GetSessionsWithPhoneNumber()
	if err != nil {
		fmt.Printf("Failed to fetch connected sessions for reconnect: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("ReconnectAllSessions: no sessions with stored JID found")
		return
	}

	fmt.Printf("ReconnectAllSessions: found %d session(s) with stored JID\n", len(sessions))

	for _, session := range sessions {
		fmt.Printf("Reconnecting session: %s (%s) [status=%s, jid=%s]\n", session.SessionName, session.ID, session.Status, session.PhoneNumber)
		go func(id string) {
			if _, err := cm.Connect(id); err != nil {
				fmt.Printf("Failed to reconnect session %s: %v\n", id, err)
				// Optional: Update status to disconnected if reconnect fails repeatedly
			}
		}(session.ID)
	}
}
