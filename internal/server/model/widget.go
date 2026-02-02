package model

type InitWidgetRequest struct {
	VisitorID   string `json:"visitor_id"`
	Fingerprint string `json:"fingerprint"`
}

type InitWidgetResponse struct {
	Token     string `json:"token"`
	VisitorID string `json:"visitor_id"`
}
