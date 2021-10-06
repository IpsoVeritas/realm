package gorm

import (
	"testing"

	"github.com/jinzhu/gorm"
	realm "github.com/IpsoVeritas/realm"
)

type service struct {
	db          *gorm.DB
	realms      realm.RealmProvider
	controllers realm.ControllerProvider
	actions     realm.ActionProvider
	invites     realm.InviteProvider
	mandates    realm.IssuedMandateProvider
	roles       realm.RoleProvider
}

func newService(t *testing.T, dbLog bool) *service {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.LogMode(dbLog)

	realms, err := NewGormRealmService(db)
	if err != nil {
		t.Fatal(err)
	}

	controllers, err := NewGormControllerService(db)
	if err != nil {
		t.Fatal(err)
	}

	actions, err := NewGormActionService(db)
	if err != nil {
		t.Fatal(err)
	}

	invites, err := NewGormInviteService(db)
	if err != nil {
		t.Fatal(err)
	}

	mandates, err := NewGormMandateService(db)
	if err != nil {
		t.Fatal(err)
	}

	roles, err := NewGormRoleService(db)
	if err != nil {
		t.Fatal(err)
	}

	svc := &service{
		db:          db,
		realms:      realms,
		controllers: controllers,
		actions:     actions,
		invites:     invites,
		mandates:    mandates,
		roles:       roles,
	}
	return svc
}
