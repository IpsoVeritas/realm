package gorm

import (
	"encoding/json"

	realm "github.com/IpsoVeritas/realm"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// GormInviteService provider using a database
type GormInviteService struct {
	db *gorm.DB
}

type inviteData struct {
	ID    string `gorm:"primary_key"`
	Realm string `gorm:"index"`
	Role  string `gorm:"index"`
	Data  []byte
}

func (inviteData) TableName() string {
	return "invites"
}

func NewGormInviteService(db *gorm.DB) (realm.InviteProvider, error) {
	p := &GormInviteService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormInviteService) Migrate() error {
	return p.db.AutoMigrate(&inviteData{}).Error
}

func (p *GormInviteService) List(realmID string) ([]*realm.Invite, error) {
	invites := make([]*inviteData, 0)
	err := p.db.Where("realm = ?", realmID).Find(&invites).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.Invite, 0)
	for _, cd := range invites {
		c := &realm.Invite{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormInviteService) ListForRole(realmID, role string) ([]*realm.Invite, error) {
	invites := make([]*inviteData, 0)
	err := p.db.Where("realm = ? AND role = ?", realmID, role).Find(&invites).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.Invite, 0)
	for _, cd := range invites {
		c := &realm.Invite{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormInviteService) Get(realmID, id string) (*realm.Invite, error) {
	ad := &inviteData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&ad).Error
	if err != nil {
		return nil, err
	}

	var c *realm.Invite
	err = json.Unmarshal(ad.Data, &c)
	c.Realm = ad.Realm

	return c, err
}

func (p *GormInviteService) Set(realmID string, c *realm.Invite) error {
	if c.ID == "" {
		c.ID = uuid.NewV4().String()
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	ad := &inviteData{
		ID:    c.ID,
		Realm: realmID,
		Role:  c.Role,
		Data:  bytes,
	}

	err = p.db.Save(&ad).Error

	return err
}

func (p *GormInviteService) Delete(realmID, id string) error {
	return p.db.Delete(&inviteData{}, "id = ? AND realm = ?", id, realmID).Error
}
