package document

import (
	"strings"
	"time"

	jose "gopkg.in/square/go-jose.v1"
)

const MandateType = "mandate"

const (
	MandateActive  = iota
	MandateRevoked = iota
)

type Mandate struct {
	Base
	Role          string            `json:"role,omitempty"`
	Label         string            `json:"label,omitempty"`
	TTL           int               `json:"ttl,omitempty"`
	Recipient     string            `json:"recipient,omitempty"`
	RecipientName string            `json:"recipientName,omitempty"`
	RecipientPK   *jose.JsonWebKey  `json:"recipientPublicKey,omitempty"`
	RequestID     string            `json:"requestId,omitempty"`
	Sender        string            `json:"sender,omitempty"`
	Params        map[string]string `json:"params,omitempty"`
}

// use this so that we can change it in the future.
func RealmRoleFormat(roleName string, realmName string) string {
	return roleName + "@" + realmName
}

func RealmRoleParse(realmRole string) (role string, realm string) {
	parts := strings.Split(realmRole, "@")
	role = parts[0]
	if len(parts) > 1 {
		realm = parts[1]
	} else {
		realm = ""
	}
	return role, realm
}

func NewMandate(role string) *Mandate {
	return &Mandate{
		Base: Base{
			Context:   Context,
			Type:      MandateType,
			Timestamp: time.Now(),
		},
		Role: role,
	}
}
