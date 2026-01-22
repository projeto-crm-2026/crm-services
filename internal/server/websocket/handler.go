package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type IncomingMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	VisitorID string `json:"visitor_id,omitempty"`
}

type MessageSaver interface {
	SaveMessage(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) error
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, chatID uint, userID *uint, visitorID string, saver MessageSaver) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Hub:       hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		ChatID:    chatID,
		UserID:    userID,
		VisitorID: visitorID,
	}

	hub.register <- client

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		var msg IncomingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			return
		}

		if saver != nil && msg.Content != "" {
			if err := saver.SaveMessage(context.Background(), chatID, userID, visitorID, msg.Content); err != nil {
				hub.logger.Error("failed to save message")
			}
		}

		hub.BroadcastToChat(chatID, message)
	})
}
