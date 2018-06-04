package document

import (
	"fmt"
	"sync"
	"time"
)

const Context = "https://brickchain.com/schema"
const BaseType = "base"

type Base struct {
	Context          string    `json:"@context,omitempty"`
	Type             string    `json:"@type"`
	SubType          string    `json:"@subtype,omitempty"`
	Timestamp        time.Time `json:"@timestamp"`
	ID               string    `json:"@id,omitempty"`
	Links            []Link    `json:"@links,omitempty"`
	Owners           []string  `json:"@owners,omitempty"`
	Callbacks        []string  `json:"@callbacks,omitempty"`
	CertificateChain string    `json:"@certificateChain,omitempty"`
	Realm            string    `json:"@realm,omitempty"`
	mu               *sync.Mutex
}

func NewBase() *Base {
	return &Base{
		Context:   Context,
		Type:      BaseType,
		Timestamp: time.Now().UTC(),
		mu:        new(sync.Mutex),
	}
}

func (b *Base) Expand() string {
	if b.Context != "" {
		return fmt.Sprintf("%s/%s", b.Context, b.Type)
	}
	return b.Type
}

func (b *Base) AddLink(link Link) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Links = append(b.Links, link)
}

func (b *Base) AddOwner(owner string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Owners = append(b.Owners, owner)
}

func (b *Base) AddCallback(uri string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Callbacks = append(b.Callbacks, uri)
}

func (b *Base) GetCertificateChain() string {
	return b.CertificateChain
}

func (b *Base) GetType() string {
	return b.Type
}

func (b *Base) GetSubType() string {
	return b.SubType
}

type Link struct {
	Type string `json:"@type"`
	ID   string `json:"@id"`
}

func NewLink(docType, id string) Link {
	return Link{
		Type: docType,
		ID:   id,
	}
}
