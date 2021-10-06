package realm

import (
	"github.com/IpsoVeritas/document"
	jose "gopkg.in/square/go-jose.v1"
)

type Realm struct {
	ID                   string                    `json:"@id,omitempty"`
	Type                 string                    `json:"@type,omitempty"`
	Label                string                    `json:"label,omitempty"`
	PublicKey            *jose.JsonWebKey          `json:"publicKey,omitempty"`
	URI                  string                    `json:"uri,omitempty"`
	GuestMandateTicketID string                    `json:"guestMandateTicketId"`
	Descriptor           *document.RealmDescriptor `json:"realmDescriptor,omitempty"`
	SignedDescriptor     string                    `json:"signedDescriptor,omitempty"`
	AdminRoles           []string                  `json:"adminRoles,omitempty"`
	OwnerRealm           bool                      `json:"ownerRealm,omitempty"`
	GuestRole            string                    `json:"guestRole,omitempty"`
}

type RealmProvider interface {
	List() ([]*Realm, error)
	Get(id string) (*Realm, error)
	Set(*Realm) error
	Delete(id string) error
}

type RealmEvent struct {
	Type  string `json:"type"`
	Realm string `json:"realm"`
}
