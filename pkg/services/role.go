package services

import (
	document "github.com/Brickchain/go-document.v2"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
)

type RoleService struct {
	p       realm.RoleProvider
	realmID string
}

func (r *RoleService) List() ([]*document.Role, error) {
	return r.p.List(r.realmID)
}

func (r *RoleService) Get(id string) (*document.Role, error) {
	return r.p.Get(r.realmID, id)
}

func (r *RoleService) ByName(name string) (*document.Role, error) {
	return r.p.ByName(r.realmID, name)
}

func (r *RoleService) Set(role *document.Role) error {
	return r.p.Set(r.realmID, role)
}

func (r *RoleService) Delete(id string) error {
	return r.p.Delete(r.realmID, id)
}
