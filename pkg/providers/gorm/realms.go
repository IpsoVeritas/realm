package gorm

import (
	"encoding/json"
	"errors"

	stats "github.com/Brickchain/go-stats.v1"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	realm "github.com/Brickchain/realm"
)

// Realm provider using a database
type GormRealmService struct {
	db *gorm.DB
}

type realmData struct {
	ID   string `gorm:"primary_key"`
	Data []byte
}

func (realmData) TableName() string {
	return "realms"
}

func NewGormRealmService(db *gorm.DB) (realm.RealmProvider, error) {
	p := &GormRealmService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormRealmService) Migrate() error {
	return p.db.AutoMigrate(&realmData{}).Error
}

func (p *GormRealmService) List() ([]*realm.Realm, error) {
	total := stats.StartTimer("services.realm.List.total")
	defer total.Stop()

	realms := make([]*realmData, 0)
	err := p.db.Find(&realms).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.Realm, 0)
	for _, rd := range realms {
		realm := &realm.Realm{}
		if err := json.Unmarshal(rd.Data, &realm); err != nil {
			return nil, errors.New("Failed to unmarshal realm data")
		}
		out = append(out, realm)
	}
	return out, nil
}

func (p *GormRealmService) Get(id string) (*realm.Realm, error) {
	total := stats.StartTimer("services.realm.Get.total")
	defer total.Stop()

	r := &realmData{}
	err := p.db.Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}

	var realm *realm.Realm
	err = json.Unmarshal(r.Data, &realm)

	return realm, err
}

func (p *GormRealmService) Set(r *realm.Realm) error {
	total := stats.StartTimer("services.realm.Set.total")
	defer total.Stop()

	if r.ID == "" {
		r.ID = uuid.Must(uuid.NewV4()).String()
	}

	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	rd := &realmData{
		ID:   r.ID,
		Data: bytes,
	}

	return p.db.Save(rd).Error
}

func (p *GormRealmService) Delete(id string) error {
	total := stats.StartTimer("services.realm.Delete.total")
	defer total.Stop()

	return p.db.Delete(&realmData{}, "id = ?", id).Error
}
