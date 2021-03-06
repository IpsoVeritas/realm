package gorm

import (
	"fmt"

	realm "github.com/IpsoVeritas/realm"
	"github.com/jinzhu/gorm"
)

// GormSettingService provider using a database
type GormSettingService struct {
	db *gorm.DB
}

type setting struct {
	ID    string `gorm:"primary_key"`
	Realm string `gorm:"index"`
	Key   string `gorm:"index"`
	Value string
}

func NewGormSettingService(db *gorm.DB) (realm.SettingProvider, error) {
	p := &GormSettingService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormSettingService) Migrate() error {
	return p.db.AutoMigrate(&setting{}).Error
}

func (p *GormSettingService) List(realmID string) ([]*realm.Setting, error) {
	settings := make([]*setting, 0)
	err := p.db.Where("realm = ?", realmID).Find(&settings).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.Setting, 0)
	for _, s := range settings {
		out = append(out, &realm.Setting{
			Realm: s.Realm,
			Key:   s.Key,
			Value: s.Value,
		})
	}

	return out, nil
}

func (p *GormSettingService) key(realmID, key string) string {
	return fmt.Sprintf("%s_%s", realmID, key)
}

func (p *GormSettingService) Get(realmID, key string) (string, error) {
	setting := &setting{}
	err := p.db.Where("id = ? AND realm = ?", p.key(realmID, key), realmID).First(&setting).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

func (p *GormSettingService) Set(realmID, key, value string) error {
	s := &setting{
		ID:    p.key(realmID, key),
		Realm: realmID,
		Key:   key,
		Value: value,
	}

	return p.db.Save(&s).Error
}

func (p *GormSettingService) Delete(realmID, key string) error {
	return p.db.Delete(&setting{}, "id = ? AND realm = ?", p.key(realmID, key), realmID).Error
}
