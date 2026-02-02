package websocket

import (
	"log/slog"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	chatRooms  map[uint]map[*Client]bool // chatID -> clients
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *slog.Logger
}

type BroadcastMessage struct {
	ChatID  uint
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		chatRooms:  make(map[uint]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     slog.Default(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.chatRooms[client.ChatID]; !ok {
				h.chatRooms[client.ChatID] = make(map[*Client]bool)
			}
			h.chatRooms[client.ChatID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.chatRooms[client.ChatID], client)
				close(client.Send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.Lock()
			if clients, ok := h.chatRooms[msg.ChatID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Message:
					default:
						close(client.Send)
						delete(h.clients, client)
						delete(h.chatRooms[msg.ChatID], client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastToChat(chatID uint, message []byte) {
	h.broadcast <- &BroadcastMessage{
		ChatID:  chatID,
		Message: message,
	}
}
