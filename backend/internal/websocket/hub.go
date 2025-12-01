package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Hub       *Hub
	SessionID string
	Conn      *websocket.Conn
	Send      chan []byte
}

type Hub struct {
	// Registered clients map[sessionID]map[*Client]bool
	Clients    map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	mu         sync.RWMutex
}

type Message struct {
	SessionID string      `json:"-"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.SessionID] == nil {
				h.Clients[client.SessionID] = make(map[*Client]bool)
			}
			h.Clients[client.SessionID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Clients[client.SessionID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.Clients, client.SessionID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			if clients, ok := h.Clients[message.SessionID]; ok {
				msgBytes, _ := json.Marshal(message)
				for client := range clients {
					select {
					case client.Send <- msgBytes:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToSession(sessionID string, msgType string, data interface{}) {
	h.Broadcast <- Message{
		SessionID: sessionID,
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for message := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, message)
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func originAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return true
	}
	for _, o := range allowed {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, sessionID string, allowedOrigins []string) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return originAllowed(origin, allowedOrigins)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{Hub: hub, SessionID: sessionID, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
