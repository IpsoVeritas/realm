package realm

import document "github.com/Brickchain/go-document.v2"

// type Role struct {
// 	ID          string `json:"@id,omitempty" gorm:"primary_key"`
// 	Type        string `json:"@type,omitempty"`
// 	Realm       string `json:"realm" gorm:"index"`
// 	Name        string `json:"name,omitempty" gorm:"index"`
// 	Description string `json:"description,omitempty"`
// 	Internal    bool   `json:"internal,omitempty"`
// 	KeyLevel    int    `json:"keyLevel,omitempty"`
// }

type RoleProvider interface {
	List(realmID string) ([]*document.Role, error)
	Get(realmID, id string) (*document.Role, error)
	ByName(realmID, name string) (*document.Role, error)
	Set(realmID string, role *document.Role) error
	Delete(realmID, id string) error
}
