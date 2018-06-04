package gorm

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	stats "gitlab.brickchain.com/libs/go-stats.v1"
)

// GormActionService provider using a database
type GormActionService struct {
	db *gorm.DB
}

type actionData struct {
	ID         string `gorm:"primary_key"`
	Realm      string `gorm:"index"`
	Controller string `gorm:"index"`
	Data       []byte
}

func (actionData) TableName() string {
	return "actions"
}

func NewGormActionService(db *gorm.DB) (realm.ActionProvider, error) {
	p := &GormActionService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormActionService) Migrate() error {
	return p.db.AutoMigrate(&actionData{}).Error
}

func (p *GormActionService) List(realmID string) ([]*realm.ControllerAction, error) {
	total := stats.StartTimer("services.action.List.total")
	defer total.Stop()

	actions := make([]*actionData, 0)
	err := p.db.Where("realm = ?", realmID).Find(&actions).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.ControllerAction, 0)
	for _, cd := range actions {
		c := &realm.ControllerAction{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormActionService) ListForController(realmID, controllerID string) ([]*realm.ControllerAction, error) {
	total := stats.StartTimer("services.action.ListForController.total")
	defer total.Stop()

	actions := make([]*actionData, 0)
	err := p.db.Where("realm = ? AND controller = ?", realmID, controllerID).Find(&actions).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.ControllerAction, 0)
	for _, cd := range actions {
		c := &realm.ControllerAction{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormActionService) Get(realmID, id string) (*realm.ControllerAction, error) {
	total := stats.StartTimer("services.action.Get.total")
	defer total.Stop()

	ad := &actionData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&ad).Error
	if err != nil {
		return nil, err
	}

	var c *realm.ControllerAction
	err = json.Unmarshal(ad.Data, &c)
	c.Realm = ad.Realm

	return c, err
}

func (p *GormActionService) Set(realmID string, c *realm.ControllerAction) error {
	total := stats.StartTimer("services.action.Set.total")
	defer total.Stop()

	if c.ID == "" {
		c.ID = uuid.Must(uuid.NewV4()).String()
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	ad := &actionData{
		ID:         c.ID,
		Realm:      realmID,
		Controller: c.ControllerID,
		Data:       bytes,
	}

	err = p.db.Save(&ad).Error

	return err
}

func (p *GormActionService) Delete(realmID, id string) error {
	total := stats.StartTimer("services.action.Delete.total")
	defer total.Stop()

	return p.db.Delete(&actionData{}, "id = ? AND realm = ?", id, realmID).Error
}
