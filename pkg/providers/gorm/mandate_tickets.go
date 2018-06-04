package gorm

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	stats "gitlab.brickchain.com/libs/go-stats.v1"
)

// GormMandateTicketService provider using a database
type GormMandateTicketService struct {
	db *gorm.DB
}

type mandateTicketData struct {
	ID    string `gorm:"primary_key"`
	Realm string `gorm:"index"`
	Data  []byte
}

func (mandateTicketData) TableName() string {
	return "mandatetickets"
}

func NewGormMandateTicketService(db *gorm.DB) (realm.MandateTicketProvider, error) {
	p := &GormMandateTicketService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormMandateTicketService) Migrate() error {
	return p.db.AutoMigrate(&mandateTicketData{}).Error
}

func (p *GormMandateTicketService) List(realmID string) ([]*realm.MandateTicket, error) {
	total := stats.StartTimer("services.mandateTicket.List.total")
	defer total.Stop()

	mandateTickets := make([]*mandateTicketData, 0)
	err := p.db.Where("realm = ?", realmID).Find(&mandateTickets).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.MandateTicket, 0)
	for _, cd := range mandateTickets {
		c := &realm.MandateTicket{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormMandateTicketService) Get(realmID, id string) (*realm.MandateTicket, error) {
	total := stats.StartTimer("services.mandateTicket.Get.total")
	defer total.Stop()

	ad := &mandateTicketData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&ad).Error
	if err != nil {
		return nil, err
	}

	var c *realm.MandateTicket
	err = json.Unmarshal(ad.Data, &c)
	c.Realm = ad.Realm

	return c, err
}

func (p *GormMandateTicketService) Set(realmID string, c *realm.MandateTicket) error {
	total := stats.StartTimer("services.mandateTicket.Set.total")
	defer total.Stop()

	if c.ID == "" {
		c.ID = uuid.Must(uuid.NewV4()).String()
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	ad := &mandateTicketData{
		ID:    c.ID,
		Realm: realmID,
		Data:  bytes,
	}

	err = p.db.Save(&ad).Error

	return err
}

func (p *GormMandateTicketService) Delete(realmID, id string) error {
	total := stats.StartTimer("services.mandateTicket.Delete.total")
	defer total.Stop()

	return p.db.Delete(&mandateTicketData{}, "id = ? AND realm = ?", id, realmID).Error
}
