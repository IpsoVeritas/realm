package document

import (
	"sync"
	"time"
)

const RevocationChecksumType = "revocation-checksum"

type RevocationChecksum struct {
	Base
	Multihash string `json:"multihash"` // The signature from this document is to be revoked
}

func NewRevocationChecksum(multihash string) *RevocationChecksum {
	r := &RevocationChecksum{
		Base: Base{
			Context:   Context,
			Type:      RevocationChecksumType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		Multihash: multihash,
	}
	return r
}
