package gorm

import (
	"encoding/json"

	realm "github.com/IpsoVeritas/realm"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// GormMandateService provider using a database
type GormMandateService struct {
	db *gorm.DB
}

type mandateData struct {
	ID    string `gorm:"primary_key"`
	Realm string `gorm:"index"`
	Role  string `gorm:"index"`
	Data  []byte
}

func (mandateData) TableName() string {
	return "mandates"
}

func NewGormMandateService(db *gorm.DB) (realm.IssuedMandateProvider, error) {
	p := &GormMandateService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormMandateService) Migrate() error {
	return p.db.AutoMigrate(&mandateData{}).Error
}

func (p *GormMandateService) List(realmID string) ([]*realm.IssuedMandate, error) {
	mandates := make([]*mandateData, 0)
	err := p.db.Where("realm = ?", realmID).Find(&mandates).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.IssuedMandate, 0)
	for _, cd := range mandates {
		c := &realm.IssuedMandate{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormMandateService) ListForRole(realmID, role string) ([]*realm.IssuedMandate, error) {
	mandates := make([]*mandateData, 0)
	err := p.db.Where("realm = ? AND role = ?", realmID, role).Find(&mandates).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.IssuedMandate, 0)
	for _, cd := range mandates {
		c := &realm.IssuedMandate{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormMandateService) Get(realmID, id string) (*realm.IssuedMandate, error) {
	ad := &mandateData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&ad).Error
	if err != nil {
		return nil, err
	}

	var c *realm.IssuedMandate
	err = json.Unmarshal(ad.Data, &c)
	c.Realm = ad.Realm

	return c, err
}

func (p *GormMandateService) Set(realmID string, c *realm.IssuedMandate) error {
	if c.ID == "" {
		c.ID = uuid.NewV4().String()
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	ad := &mandateData{
		ID:    c.ID,
		Realm: realmID,
		Role:  c.Role,
		Data:  bytes,
	}

	err = p.db.Save(&ad).Error

	return err
}

func (p *GormMandateService) Delete(realmID, id string) error {
	return p.db.Delete(&mandateData{}, "id = ? AND realm = ?", id, realmID).Error
}
