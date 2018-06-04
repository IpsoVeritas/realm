package realm

type Setting struct {
	Realm string
	Key   string
	Value string
}

type SettingProvider interface {
	List(realmID string) ([]*Setting, error)
	Get(realmID, key string) (string, error)
	Set(realmID, key, value string) error
	Delete(realmID, key string) error
}
