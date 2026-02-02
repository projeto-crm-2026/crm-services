package model

type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewSuccessResponse(message string, data any) APIResponse {
	return APIResponse{
		Message: message,
		Data:    data,
	}
}
