package gorm

import (
	"encoding/json"

	stats "github.com/Brickchain/go-stats.v1"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	realm "github.com/Brickchain/realm"
)

// GormControllerService provider using a database
type GormControllerService struct {
	db *gorm.DB
}

type controllerData struct {
	ID       string `gorm:"primary_key"`
	Realm    string `gorm:"index"`
	Priority int
	Data     []byte
}

func (controllerData) TableName() string {
	return "controllers"
}

func NewGormControllerService(db *gorm.DB) (realm.ControllerProvider, error) {
	p := &GormControllerService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormControllerService) Migrate() error {
	return p.db.AutoMigrate(&controllerData{}).Error
}

func (p *GormControllerService) List(realmID string) ([]*realm.Controller, error) {
	total := stats.StartTimer("services.controller.List.total")
	defer total.Stop()

	controllers := make([]*controllerData, 0)
	err := p.db.Where("realm = ?", realmID).Order("priority desc").Find(&controllers).Error
	if err != nil {
		return nil, err
	}

	out := make([]*realm.Controller, 0)
	for _, cd := range controllers {
		c := &realm.Controller{}
		err = json.Unmarshal(cd.Data, &c)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *GormControllerService) Get(realmID, id string) (*realm.Controller, error) {
	total := stats.StartTimer("services.controller.Get.total")
	defer total.Stop()

	cd := &controllerData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&cd).Error
	if err != nil {
		return nil, err
	}

	var c *realm.Controller
	err = json.Unmarshal(cd.Data, &c)
	c.Realm = cd.Realm
	c.Priority = cd.Priority

	return c, err
}

func (p *GormControllerService) Set(realmID string, c *realm.Controller) error {
	total := stats.StartTimer("services.controller.Set.total")
	defer total.Stop()

	if c.ID == "" {
		c.ID = uuid.Must(uuid.NewV4()).String()
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	cd := &controllerData{
		ID:       c.ID,
		Realm:    realmID,
		Priority: c.Priority,
		Data:     bytes,
	}

	err = p.db.Save(&cd).Error

	return err
}

func (p *GormControllerService) Delete(realmID, id string) error {
	total := stats.StartTimer("services.controller.Delete.total")
	defer total.Stop()

	return p.db.Delete(&controllerData{}, "id = ? AND realm = ?", id, realmID).Error
}
