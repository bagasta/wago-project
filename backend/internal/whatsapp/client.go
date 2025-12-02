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

func (cm *ClientManager) Connect(sessionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.Clients[sessionID]; ok {
		return nil // Already connected
	}

	// Get device store
	// Actually whatsmeow uses JID (phone number) to identify devices in the store.
	// But we want to map our UUID sessionID to a whatsmeow device.
	// sqlstore.GetDevice(jid) gets a device by JID.
	// If we want to support multiple sessions, we need to manage the mapping.
	// However, whatsmeow's sqlstore is designed to hold multiple devices.
	// We can use `NewDevice` to create a new one if it doesn't exist, but we need to know the JID first?
	// No, `GetFirstDevice` gets the first one.
	// We need a way to link our sessionID to the whatsmeow device.
	// A common pattern is to use a separate store per session (SQLite) or use the 'ID' column in whatsmeow_device table if possible.
	// But `sqlstore` manages the schema.

	// Alternative: We can just use the sessionID as the "ID" if we were using a custom store, but here we are using the standard one.
	// Let's try to fetch all devices and see if we can map them.
	// Or better: Since we are starting fresh, we can create a NEW device for this session.
	// But `container.NewDevice()` returns a new device. We need to persist which device belongs to which session.
	// We should probably store the `JID` in our `sessions` table after login.
	// But before login, we don't have a JID.

	// Wait, `whatsmeow` documentation says:
	// "If you want to have multiple sessions, you should use a Container."
	// "The Container will manage the database connection and allow you to get Device stores."
	// We can use `container.GetDevice(jid)` but we don't have JID yet.
	// We can use `container.NewDevice()` to create a new one.
	// BUT, how do we retrieve it later by `sessionID`?
	// We need to store the mapping `sessionID` -> `JID` (or `DeviceID` internal to whatsmeow).
	// Actually, `whatsmeow` doesn't expose an internal ID easily other than JID.

	// WORKAROUND for this specific requirement (Multi-session by UUID):
	// We can't easily use the shared SQLStore if we don't know the JID.
	// A common approach is to use a separate SQLite file per session, OR use the Postgres store but we need to know the JID.
	// Since we don't know the JID before QR scan, we have a "chicken and egg" problem with `GetDevice(jid)`.

	// SOLUTION: Use `container.NewDevice()` which creates a new device in the DB.
	// It returns a `*store.Device`.
	// But wait, `NewDevice()` doesn't take an ID. It generates one? No, it expects us to have a JID eventually.
	// Actually, `whatsmeow` stores devices keyed by JID.
	// If we are "pre-login", we don't have a JID.
	// `whatsmeow` handles this by allowing a device to be created without a JID, and then it gets updated upon login?
	// No, `NewDevice()` creates a blank device.
	// We need to persist the association between `sessionID` and the device.
	// But `whatsmeow`'s `Device` struct doesn't have an external ID we can set.

	// Let's look at `whatsmeow` source or common patterns.
	// Usually, people use a simple file store for each session: `session-UUID.db`.
	// Since we are using Postgres, we are forced to use `sqlstore`.
	// `sqlstore` has `GetDevice(jid)` and `GetAllDevices()`.
	// If we use `NewDevice()`, it creates a device. We can use it to login.
	// Once logged in, it has a JID.
	// But what if we restart the server? We need to load the SAME device for the SAME sessionID.
	// We can't rely on JID because we might not be logged in yet (just QR stage).

	// OK, the robust way for multi-session with UUIDs is to use a separate database/schema or just use SQLite files.
	// Given the constraints (Postgres), maybe we can hack it?
	// OR, we can just use `NewDevice()` and if we restart, we lose the "pre-login" sessions?
	// That's acceptable for "QR stage".
	// But for "Connected" sessions, we MUST recover them.
	// Connected sessions have a JID (Phone Number).
	// We have `phone_number` in our `sessions` table!
	// So:
	// 1. If session has `phone_number`, we call `container.GetDevice(parsedJID)`.
	// 2. If session has NO `phone_number`, we call `container.NewDevice()`.
	//    - If we restart server while in QR mode, the user has to scan again. This is acceptable.

	// Let's proceed with this logic.

	var deviceStore *store.Device
	session, err := cm.SessionRepo.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
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
								cm.SessionRepo.UpdateSessionStatus(sessionID, session.Status, dev.ID.String(), session.DeviceInfo)
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
			return err
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
					cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusQR, "", nil)
				} else {
					// Timeout or success?
					// Success is handled by EventHandler
				}
			}
		}()
	} else {
		// Already logged in
		err = client.Connect()
		if err != nil {
			return err
		}
		// Update status just in case
		// cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusConnected, client.Store.ID.User, nil)
	}

	return nil
}

func (cm *ClientManager) disconnect(sessionID string, updateStatus bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.Clients[sessionID]; ok {
		client.Disconnect()
		delete(cm.Clients, sessionID)
		if updateStatus {
			cm.SessionRepo.UpdateSessionStatus(sessionID, model.SessionStatusDisconnected, "", nil)
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
			if err := cm.Connect(id); err != nil {
				fmt.Printf("Failed to reconnect session %s: %v\n", id, err)
				// Optional: Update status to disconnected if reconnect fails repeatedly
			}
		}(session.ID)
	}
}
