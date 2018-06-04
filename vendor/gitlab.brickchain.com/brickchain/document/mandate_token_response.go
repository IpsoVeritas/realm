package document

import (
	"time"
)

const MandateTokenResponseType = "mandate-token-response"

type MandateTokenResponse struct {
	Base
	Token string `json:"token"`
}

func NewMandateTokenResponse(token string) *MandateTokenResponse {
	return &MandateTokenResponse{
		Base: Base{
			Context:   Context,
			Type:      MandateTokenResponseType,
			Timestamp: time.Now().UTC(),
		},
		Token: token,
	}
}
