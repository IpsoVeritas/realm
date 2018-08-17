package keys

import (
	hash "crypto"
	"encoding/base64"
	"encoding/json"

	"github.com/Brickchain/go-crypto.v2"
	jose "gopkg.in/square/go-jose.v1"
)

// StoredKey is the format of the stored encrypted key
type StoredKey struct {
	ID   string `json:"id" gorm:"primary_key"`
	Data string `json:"data"`
}

// NewStoredKey creates a new StoredKey object with the ID field set
func NewStoredKey(id string) *StoredKey {
	return &StoredKey{
		ID: id,
	}
}

// Decrypt decrypts the encrypted data payload of the StoredKey
func (s *StoredKey) Decrypt(kek []byte) (*jose.JsonWebKey, error) {
	jwe, err := crypto.UnmarshalJWE(s.Data)
	if err != nil {
		return nil, err
	}

	payload, err := jwe.Decrypt(kek)
	if err != nil {
		return nil, err
	}

	k := &jose.JsonWebKey{}
	if err := json.Unmarshal(payload, &k); err != nil {
		return nil, err
	}

	if k.KeyID == "" {
		keyTPbytes, _ := k.Thumbprint(hash.SHA256)
		k.KeyID = base64.URLEncoding.EncodeToString(keyTPbytes)
	}

	return k, nil
}

// Encrypt a JWK with a kek (Key Encryption Key) and store in the Data field of the StoredKey object
func (s *StoredKey) Encrypt(key *jose.JsonWebKey, kek []byte) error {
	payload, err := json.Marshal(key)
	if err != nil {
		return err
	}

	encrypter, err := crypto.NewSymmetricEncrypter(kek)
	if err != nil {
		return err
	}

	enc, err := encrypter.Encrypt(payload)
	if err != nil {
		return err
	}

	ser, err := enc.CompactSerialize()
	if err != nil {
		return err
	}

	s.Data = ser

	return nil
}
