package model

type CreateAPIKeyRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type APIKeyResponse struct {
	ID        uint   `json:"id"`
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key,omitempty"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	IsActive  bool   `json:"is_active"`
}
