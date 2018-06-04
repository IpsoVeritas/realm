package document

import (
	"time"
)

const MandateTokenType = "mandate-token"

type MandateToken struct {
	Base
	Mandate  string   `json:"mandate,omitempty"`
	Mandates []string `json:"mandates,omitempty"`
	URI      string   `json:"uri,omitempty"`
	TTL      int      `json:"ttl,omitempty"`
}

func NewMandateToken(mandate string, uri string, ttl int) *MandateToken {
	return &MandateToken{
		Base: Base{
			Context:   Context,
			Type:      MandateTokenType,
			Timestamp: time.Now().UTC(),
		},
		Mandate: mandate,
		URI:     uri,
		TTL:     ttl,
	}
}
