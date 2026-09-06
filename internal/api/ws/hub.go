package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for this phase
	},
}

// Client represents a connected WebSocket client
type Client struct {
	ID         uuid.UUID
	DatabaseID uuid.UUID
	Conn       *websocket.Conn
	Send       chan []byte
	Hub        *Hub
}

// Hub manages active WebSocket clients and broadcasts messages
type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	Broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	log        zerolog.Logger
}

type Message struct {
	DatabaseID uuid.UUID
	Payload    interface{}
}

func NewHub(log zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		Broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		log:        log.With().Str("component", "ws_hub").Logger(),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.DatabaseID]; !ok {
				h.clients[client.DatabaseID] = make(map[*Client]bool)
			}
			h.clients[client.DatabaseID][client] = true
			h.mu.Unlock()
			h.log.Debug().Str("client_id", client.ID.String()).Str("db_id", client.DatabaseID.String()).Msg("Client registered")
		
		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.DatabaseID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.DatabaseID)
					}
				}
			}
			h.mu.Unlock()
			h.log.Debug().Str("client_id", client.ID.String()).Str("db_id", client.DatabaseID.String()).Msg("Client unregistered")
		
		case message := <-h.Broadcast:
			h.mu.RLock()
			clients := h.clients[message.DatabaseID]
			
			if len(clients) > 0 {
				payloadBytes, err := json.Marshal(message.Payload)
				if err != nil {
					h.log.Error().Err(err).Msg("Failed to marshal broadcast payload")
					h.mu.RUnlock()
					continue
				}

				for client := range clients {
					select {
					case client.Send <- payloadBytes:
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

func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request, dbID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to upgrade websocket connection")
		return
	}

	client := &Client{
		ID:         uuid.New(),
		DatabaseID: dbID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        h,
	}

	h.register <- client

	// Start pump routines
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
