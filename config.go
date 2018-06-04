package realm

type RealmConfig struct {
	ServicesFeed string   `json:"servicesFeed"`
	AdminRoles   []string `json:"adminRoles"`
}
