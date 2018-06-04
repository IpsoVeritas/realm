package document

import (
	"time"

	jose "gopkg.in/square/go-jose.v1"
)

const RealmDescriptorType = SchemaLocation + "/realm-descriptor.json"

type RealmDescriptor struct {
	Base
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	PublicKey   *jose.JsonWebKey `json:"publicKey,omitempty"`
	InviteURL   string           `json:"inviteURL,omitempty"`
	ServicesURL string           `json:"servicesURL,omitempty"`
	Icon        string           `json:"icon,omitempty"`
	Banner      string           `json:"banner,omitempty"`
}

func NewRealmDescriptor(name string, publicKey *jose.JsonWebKey, servicesURL string) *RealmDescriptor {
	return &RealmDescriptor{
		Base: Base{
			Type:      RealmDescriptorType,
			Timestamp: time.Now().UTC(),
		},
		Name:        name,
		PublicKey:   publicKey,
		ServicesURL: servicesURL,
	}
}
