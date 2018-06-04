package realm

import (
	"github.com/Brickchain/go-document.v2"
	jose "gopkg.in/square/go-jose.v1"
)

type Realm struct {
	ID                   string                    `json:"@id,omitempty"`
	Type                 string                    `json:"@type,omitempty"`
	Name                 string                    `json:"name,omitempty"`
	PublicKey            *jose.JsonWebKey          `json:"publicKey,omitempty"`
	Description          string                    `json:"description,omitempty"`
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
