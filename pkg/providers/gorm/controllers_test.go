package gorm

import (
	"testing"

	"github.com/Brickchain/go-document.v2"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	realm "github.com/Brickchain/realm"
)

func TestControllerService_Get(t *testing.T) {
	type test struct {
		name    string
		svc     realm.ControllerProvider
		prepare func(*testing.T, *test)
		id      string
		wantErr bool
	}
	tests := []test{
		{
			name: "Get",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Controller{
					Base: document.Base{
						ID:    "abc",
						Realm: "abc",
					},
				}
				if err := tt.svc.Set("abc", &r); err != nil {
					t.Fatal(err)
				}
			},
			id:      "abc",
			wantErr: false,
		},
		{
			name: "Get_Controller_not_exist",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Controller{
					Base: document.Base{
						ID:    "abc",
						Realm: "abc",
					},
				}
				if err := tt.svc.Set("abc", &r); err != nil {
					t.Fatal(err)
				}
			},
			id:      "fdfgfgd",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).controllers
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.Get("abc", tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ControllerService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("ControllerService.Get() = ID: %v, want ID: %v", got.ID, tt.id)
			}
		})
	}
}

func TestControllerService_List(t *testing.T) {
	type test struct {
		name    string
		svc     realm.ControllerProvider
		prepare func(*testing.T, *test)
		realm   string
		count   int
		wantErr bool
	}
	tests := []test{
		{
			name: "List",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Controller{}
				if err := tt.svc.Set(tt.realm, &r); err != nil {
					t.Fatal(err)
				}
			},
			realm:   "abc",
			count:   1,
			wantErr: false,
		},
		{
			name:    "List_empty",
			prepare: func(t *testing.T, tt *test) {},
			realm:   "abc",
			count:   0,
			wantErr: false,
		},
		{
			name: "List_another_realm",
			prepare: func(t *testing.T, tt *test) {
				r := realm.Controller{}
				if err := tt.svc.Set("abc", &r); err != nil {
					t.Fatal(err)
				}
			},
			realm:   "cde",
			count:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).controllers
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.List(tt.realm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ControllerService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.count {
				t.Errorf("ControllerService.List() = count: %d, want count: %d", len(got), tt.count)
			}
		})
	}
}
