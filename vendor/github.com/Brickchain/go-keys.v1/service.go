package keys

import "errors"

// StoredKeyService describes the methods used for storing and retrieving StoredKey objects
type StoredKeyService interface {
	Get(id string) (*StoredKey, error)
	Save(*StoredKey) error
}

// ErrNoSuchKey is what the StoredKeyService should return if no StoredKey was found
var ErrNoSuchKey = errors.New("No such key")
