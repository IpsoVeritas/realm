package document

import (
	"encoding/json"
	"sync"
	"time"

	jose "gopkg.in/square/go-jose.v1"
)

const RevocationType = "revocation"

type Revocation struct {
	Base
	Checksum *jose.JsonWebSignature `json:"checksum"` // A Multihash string
}

func NewRevocation(checksum *jose.JsonWebSignature) *Revocation {
	r := &Revocation{
		Base: Base{
			Context:   Context,
			Type:      RevocationType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		Checksum: checksum,
	}
	return r
}

func (r *Revocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Base
		Checksum string `json:"checksum"` // The signature from this document is to be revoked
	}{
		Base:     r.Base,
		Checksum: r.Checksum.FullSerialize(),
	})
}

func (r *Revocation) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Base
		Checksum string `json:"checksum"` // The signature from this document is to be revoked
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	r.Base = aux.Base
	r.Checksum, err = jose.ParseSigned(aux.Checksum)
	if err != nil {
		return err
	}

	return nil
}
