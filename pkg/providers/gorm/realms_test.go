package gorm

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	realm "github.com/Brickchain/realm"
)

func TestRealmService_Get(t *testing.T) {
	type test struct {
		name    string
		svc     realm.RealmProvider
		prepare func(*testing.T, *test)
		id      string
		wantErr bool
	}
	tests := []test{
		{
			name: "Get",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Realm{ID: "abc"}
				if err := tt.svc.Set(&r); err != nil {
					t.Fatal(err)
				}
			},
			id:      "abc",
			wantErr: false,
		},
		{
			name: "Get_Realm_not_exist",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Realm{ID: "abc"}
				if err := tt.svc.Set(&r); err != nil {
					t.Fatal(err)
				}
			},
			id:      "fdfgfgd",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).realms
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.Get(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RealmService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("RealmService.Get() = ID: %v, want ID: %v", got.ID, tt.id)
			}
		})
	}
}
