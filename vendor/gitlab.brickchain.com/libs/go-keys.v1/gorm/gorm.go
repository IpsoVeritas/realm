package gorm

import (
	"github.com/jinzhu/gorm"
	keys "gitlab.brickchain.com/libs/go-keys.v1"
)

// GormStoredKeyService stores StorredKey objects in a database using the Gorm ORM
type gormStoredKeyService struct {
	db *gorm.DB
}

// NewGormStoredKeyService does database migrations and returns a new StoredKeyService using the gorm ORM
func NewGormStoredKeyService(db *gorm.DB) (keys.StoredKeyService, error) {
	db.AutoMigrate(&keys.StoredKey{})

	p := &gormStoredKeyService{
		db: db,
	}

	return p, nil
}

// Get a SigningKey from the database
func (g *gormStoredKeyService) Get(id string) (*keys.StoredKey, error) {
	sk := &keys.StoredKey{}
	count := -1
	err := g.db.Where("id = ?", id).First(&sk).Count(&count).Error
	if err != nil {
		if count == 0 {
			return nil, keys.ErrNoSuchKey
		}
		return nil, err
	}

	return sk, nil
}

// Save a StoredKey to the database
func (g *gormStoredKeyService) Save(sk *keys.StoredKey) error {
	return g.db.Save(&sk).Error
}
