package models

type LoginRequest struct {
	Address   string `json:"address" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
