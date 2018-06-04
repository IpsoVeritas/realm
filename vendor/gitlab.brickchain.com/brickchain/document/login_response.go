package document

import (
	"time"
)

const LoginResponseType = "login-response"

type LoginResponse struct {
	Base
	Chain    string   `json:"chain"`
	Mandates []string `json:"mandates"`
}

func NewLoginResponse() *LoginResponse {
	return &LoginResponse{
		Base: Base{
			Context:   Context,
			Type:      LoginResponseType,
			Timestamp: time.Now().UTC(),
		},
	}
}
