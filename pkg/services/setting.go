package services

import realm "gitlab.brickchain.com/brickchain/realm-ng"

type SettingService struct {
	p       realm.SettingProvider
	realmID string
}

func (r *SettingService) List() ([]*realm.Setting, error) {
	return r.p.List(r.realmID)
}

func (r *SettingService) Get(key string) (string, error) {
	return r.p.Get(r.realmID, key)
}

func (r *SettingService) Set(key, value string) error {
	return r.p.Set(r.realmID, key, value)
}

func (r *SettingService) Delete(key string) error {
	return r.p.Delete(r.realmID, key)
}
