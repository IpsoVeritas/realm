package services

import (
	"encoding/json"
	"fmt"

	"github.com/Brickchain/go-document.v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
)

type ActionService struct {
	base             string
	bootstrapRealmID string
	p                realm.ActionProvider
	realmID          string
	realm            *RealmService
}

func (a *ActionService) List() ([]*realm.ControllerAction, error) {
	return a.p.List(a.realmID)
}

func (a *ActionService) Get(id string) (*realm.ControllerAction, error) {
	return a.p.Get(a.realmID, id)
}

func (a *ActionService) Set(action *realm.ControllerAction) error {
	action.Realm = a.realmID
	return a.p.Set(a.realmID, action)
}

func (a *ActionService) Delete(id string) error {
	return a.p.Delete(a.realmID, id)
}

func (a *ActionService) ListForController(controllerID string) ([]*realm.ControllerAction, error) {
	return a.p.ListForController(a.realmID, controllerID)
}

func (a *ActionService) Services(mandates []*document.Mandate) (*document.Multipart, error) {
	realmData, err := a.realm.Realm()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get realm")
	}

	descriptors := make([]*realm.ControllerAction, 0)

	if len(mandates) > 0 {
		list, err := a.List()
		if err != nil {
			return nil, errors.Wrap(err, "failed to list actions")
		}

		for _, mandate := range mandates {
			isAdmin := false
			for _, adminRole := range realmData.AdminRoles {
				if adminRole == mandate.Role {
					isAdmin = true
				}
			}
			if isAdmin {
				loginAction := document.NewActionDescriptor("Manage your service place", realmData.AdminRoles, 1000, a.base)
				loginAction.ID = fmt.Sprintf("%s-admin", realmData.ID)
				loginAction.UIURI = fmt.Sprintf("%s?realm=%s", viper.GetString("admin_url"), realmData.ID)
				loginAction.Icon = "https://actions.brickchain.com/booking-webapp/release/icons/information.png"
				loginAction.Interfaces = []string{
					"https://interfaces.brickchain.com/v1/realm-admin.json",
				}
				loginAction.Params = map[string]string{
					"backend": fmt.Sprintf("%s/realm/v2", a.base),
				}

				if a.realmID == a.bootstrapRealmID {
					loginAction.Params["createRealms"] = "true"
				}

				descriptors = append(descriptors, &realm.ControllerAction{
					ActionDescriptor: *loginAction,
				})
			}

			for _, action := range list {
				hasRole := false
				if action.Roles != nil {
					for _, role := range action.Roles {
						if role == mandate.Role {
							hasRole = true
						}
					}
				}

				if hasRole {
					descriptors = append(descriptors, action)
				}
			}
		}
	} else {
		loginAction := document.NewActionDescriptor("Manage your service place", realmData.AdminRoles, 1000, "")
		loginAction.ID = fmt.Sprintf("%s-admin", realmData.ID)
		loginAction.Interfaces = []string{
			"https://interfaces.brickchain.com/v1/realm-admin.json",
		}
		loginAction.Params = map[string]string{
			"backend":        fmt.Sprintf("%s/realm/v2", a.base),
			"proxy_endpoint": viper.GetString("proxy_endpoint"),
		}

		if a.realmID == a.bootstrapRealmID {
			loginAction.Params["createRealms"] = "true"
		}

		descriptors = append(descriptors, &realm.ControllerAction{
			ActionDescriptor: *loginAction,
		})

		publicRole, err := a.realm.Settings().Get("publicRole")
		if err == nil && publicRole != "" {
			joinAction := document.NewActionDescriptor("Join realm", []string{}, 1000, fmt.Sprintf("%s/realm/v2/realms/%s/do/join", a.base, realmData.ID))
			joinAction.ID = fmt.Sprintf("%s-join", realmData.ID)
			joinAction.Interfaces = []string{
				"https://interfaces.brickchain.com/v1/public-role.json",
			}

			descriptors = append(descriptors, &realm.ControllerAction{
				ActionDescriptor: *joinAction,
			})
		}
	}

	mp := document.NewMultipart()

	for _, desc := range descriptors {
		if desc.Signed == "" {
			descBytes, err := json.Marshal(desc.ActionDescriptor)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal descriptor")
			}
			descSigned, err := a.realm.Sign(descBytes)
			if err != nil {
				return nil, errors.Wrap(err, "failed to sign descriptor")
			}
			desc.Signed, err = descSigned.CompactSerialize()
			if err != nil {
				return nil, errors.Wrap(err, "failed to serialize JWS")
			}
		}
		mp.Append(document.Part{
			Name:     desc.ID,
			Encoding: "text/plain+jws",
			Document: desc.Signed,
		})
	}

	return mp, nil
}
