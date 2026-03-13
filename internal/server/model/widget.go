package model

type InitWidgetRequest struct {
	VisitorID   string `json:"visitor_id"`
	Fingerprint string `json:"fingerprint"`
	ChatID      *uint  `json:"chat_id,omitempty"`
}

type InitWidgetResponse struct {
	VisitorID string        `json:"visitor_id"`
	Chat      *ChatResponse `json:"chat,omitempty"`
}
