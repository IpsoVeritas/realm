package gorm

import (
	"testing"

	document "github.com/Brickchain/go-document.v2"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	realm "github.com/Brickchain/realm"
)

func TestRoleService_Get(t *testing.T) {
	type test struct {
		name    string
		svc     realm.RoleProvider
		prepare func(*testing.T, *test)
		id      string
		wantErr bool
	}
	tests := []test{
		{
			name: "Get",
			prepare: func(t *testing.T, tt *test) {
				r := document.NewRole("test@abc")
				if err := tt.svc.Set("abc", r); err != nil {
					t.Fatal(err)
				}
				tt.id = r.ID
			},
			wantErr: false,
		},
		{
			name: "Get_Role_not_exist",
			prepare: func(t *testing.T, tt *test) {
				r := document.NewRole("test@abc")
				if err := tt.svc.Set("abc", r); err != nil {
					t.Fatal(err)
				}
			},
			id:      "fdfgfgd",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).roles
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.Get("abc", tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("RoleService.Get() = ID: %v, want ID: %v", got.ID, tt.id)
			}
		})
	}
}

func TestRoleService_List(t *testing.T) {
	type test struct {
		name    string
		svc     realm.RoleProvider
		prepare func(*testing.T, *test)
		realm   string
		count   int
		wantErr bool
	}
	tests := []test{
		{
			name: "List",
			prepare: func(t *testing.T, tt *test) {
				r := document.NewRole("test@abc")
				if err := tt.svc.Set(tt.realm, r); err != nil {
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
				r := document.NewRole("test@abc")
				if err := tt.svc.Set("abc", r); err != nil {
					t.Fatal(err)
				}
			},
			realm:   "cde",
			count:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).roles
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.List(tt.realm)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.count {
				t.Errorf("RoleService.List() = count: %d, want count: %d", len(got), tt.count)
			}
		})
	}
}

func TestRoleService_ByName(t *testing.T) {
	type test struct {
		name    string
		svc     realm.RoleProvider
		prepare func(*testing.T, *test)
		realm   string
		role    string
		id      string
		wantErr bool
	}
	tests := []test{
		{
			name: "Get",
			prepare: func(t *testing.T, tt *test) {
				r := document.NewRole(tt.role)
				if err := tt.svc.Set(tt.realm, r); err != nil {
					t.Fatal(err)
				}
				tt.id = r.ID
			},
			realm:   "abc",
			role:    "test@example.com",
			wantErr: false,
		},
		{
			name:    "Get_not_exists",
			prepare: func(t *testing.T, tt *test) {},
			realm:   "abc",
			role:    "test@example.com",
			wantErr: true,
		},
		{
			name: "List_another_role",
			prepare: func(t *testing.T, tt *test) {
				r := document.NewRole("admin@example.com")
				if err := tt.svc.Set(tt.realm, r); err != nil {
					t.Fatal(err)
				}
				tt.id = r.ID
			},
			realm:   "abc",
			role:    "test@example.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt.svc = newService(t, false).roles
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, &tt)
			}
			got, err := tt.svc.ByName(tt.realm, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleService.ByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("RoleService.ByName() = got ID: %s, want ID: %s", got.ID, tt.id)
			}
		})
	}
}
