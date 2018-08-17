package gorm

import (
	"encoding/json"

	document "github.com/Brickchain/go-document.v2"
	stats "github.com/Brickchain/go-stats.v1"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
)

type roleData struct {
	ID    string `gorm:"primary_key"`
	Realm string `gorm:"index"`
	Role  string `gorm:"index"`
	Data  []byte
}

// GormRoleService provider using a database
type GormRoleService struct {
	db *gorm.DB
}

func NewGormRoleService(db *gorm.DB) (realm.RoleProvider, error) {
	p := &GormRoleService{
		db: db,
	}

	if err := p.Migrate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *GormRoleService) Migrate() error {
	return p.db.AutoMigrate(&roleData{}).Error
}

func (p *GormRoleService) List(realmID string) ([]*document.Role, error) {
	total := stats.StartTimer("services.role.List.total")
	defer total.Stop()

	rs := make([]roleData, 0)
	err := p.db.Where("realm = ?", realmID).Find(&rs).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*document.Role, 0)
	for _, r := range rs {
		role := &document.Role{}
		if err := json.Unmarshal(r.Data, &role); err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (p *GormRoleService) ByName(realmID, name string) (*document.Role, error) {
	total := stats.StartTimer("services.role.ByName.total")
	defer total.Stop()

	r := roleData{}
	err := p.db.Where("realm = ? AND role = ?", realmID, name).First(&r).Error
	if err != nil {
		return nil, err
	}

	role := &document.Role{}
	if err := json.Unmarshal(r.Data, &role); err != nil {
		return nil, err
	}

	return role, nil
}

func (p *GormRoleService) Get(realmID, id string) (*document.Role, error) {
	total := stats.StartTimer("services.role.Get.total")
	defer total.Stop()

	r := roleData{}
	err := p.db.Where("id = ? AND realm = ?", id, realmID).First(&r).Error
	if err != nil {
		return nil, err
	}

	role := &document.Role{}
	if err := json.Unmarshal(r.Data, &role); err != nil {
		return nil, err
	}

	return role, nil
}

func (p *GormRoleService) Set(realmID string, role *document.Role) error {
	total := stats.StartTimer("services.role.Set.total")
	defer total.Stop()

	if role.ID == "" {
		role.ID = uuid.Must(uuid.NewV4()).String()
	}

	role.Realm = realmID

	r := roleData{
		ID:    role.ID,
		Role:  role.Name,
		Realm: role.Realm,
	}

	var err error
	r.Data, err = json.Marshal(role)
	if err != nil {
		return err
	}

	return p.db.Save(&r).Error
}

func (p *GormRoleService) Delete(realmID, id string) error {
	total := stats.StartTimer("services.role.Delete.total")
	defer total.Stop()

	return p.db.Delete(&roleData{}, "id = ? AND realm = ?", id, realmID).Error
}
