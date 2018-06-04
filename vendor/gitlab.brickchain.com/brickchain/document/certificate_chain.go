package document

import (
	"time"

	"strings"

	jose "gopkg.in/square/go-jose.v1"
)

const CertificateChainType = "certificate-chain"

type CertificateChain struct {
	Type             string           `json:"@type,omitempty"`
	Timestamp        time.Time        `json:"timestamp"`
	TTL              int              `json:"ttl,omitempty"`
	Root             *jose.JsonWebKey `json:"root,omitempty"`
	SubKey           *jose.JsonWebKey `json:"subKey,omitempty"`
	KeyType          string           `json:"keyType,omitempty"`
	DocumentTypes    []string         `json:"documentTypes,omitempty"`
	KeyLevel         int              `json:"keyLevel,omitempty"`
	CertificateChain string           `json:"certificateChain,omitempty"`
}

func NewCertificateChain(root, subKey *jose.JsonWebKey, keyLevel int, keyType string, ttl int) *CertificateChain {
	return &CertificateChain{
		Type:          CertificateChainType,
		Timestamp:     time.Now().UTC(),
		Root:          root,
		SubKey:        subKey,
		KeyType:       keyType,
		DocumentTypes: []string{"*"},
		TTL:           ttl,
		KeyLevel:      keyLevel,
	}
}

func (c *CertificateChain) HasExpired() bool {
	return time.Now().UTC().After(c.Timestamp.Add(time.Second * time.Duration(c.TTL)))
}

func (c *CertificateChain) AllowedType(doc BaseInterface) bool {
	for _, allowedType := range c.DocumentTypes {
		if strings.Contains(allowedType, "/") {
			parts := strings.Split(allowedType, "/")
			if len(parts) < 2 {
				return false
			}
			if parts[0] == "*" || parts[0] == doc.GetType() {
				if parts[1] == "*" || parts[1] == doc.GetSubType() {
					return true
				}
			}
		} else {
			if doc.GetType() == allowedType || allowedType == "*" {
				return true
			}
		}
	}
	return false
}
