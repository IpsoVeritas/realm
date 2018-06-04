package document

import (
	"time"

	jose "gopkg.in/square/go-jose.v1"
)

const RealmDescriptorType = "realm-descriptor"

type KeyHistoryEntry struct {
	Timestamp time.Time        `json:"timestamp"`
	Key       *jose.JsonWebKey `json:"key"`
}

type RealmDescriptor struct {
	Base
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	PublicKey   *jose.JsonWebKey  `json:"publicKey,omitempty"`
	Endpoints   map[string]string `json:"endpoints,omitempty"`
	InviteURL   string            `json:"inviteURL,omitempty"`
	ServicesURL string            `json:"servicesURL,omitempty"`
	KeyHistory  []string          `json:"keyHistory,omitempty"`
	ActionsURL  string            `json:"actionsURL,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Banner      string            `json:"banner,omitempty"`
}

func NewRealmDescriptor(name string, publicKey *jose.JsonWebKey, inviteURL string, servicesURL string) *RealmDescriptor {
	return &RealmDescriptor{
		Base: Base{
			Context:   Context,
			Type:      RealmDescriptorType,
			Timestamp: time.Now(),
		},
		Name:        name,
		PublicKey:   publicKey,
		InviteURL:   inviteURL,
		ServicesURL: servicesURL,
	}
}
