package realm

type Invite struct {
	ID          string `json:"@id,omitempty" gorm:"primary_key"`
	Type        string `json:"@type,omitempty"`
	Realm       string `json:"realm" gorm:"index"`
	Name        string `json:"name,omitempty"`
	Role        string `json:"role,omitempty" gorm:"index"`
	Status      string `json:"status,omitempty"`
	TicketID    string `json:"ticketID,omitempty"`
	MessageType string `json:"messageType,omitempty"`
	MessageURI  string `json:"messageURI,omitempty"`
	Sent        bool   `json:"sent,omitempty"`
	Text        string `json:"text,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	Sender      string `json:"sender,omitempty"`
	CreateUser  bool   `json:"createUser,omitempty"`
	KeyLevel    int    `json:"keyLevel,omitempty"`
}

type InviteProvider interface {
	List(realmID string) ([]*Invite, error)
	Get(realmID, id string) (*Invite, error)
	Set(realmID string, invite *Invite) error
	Delete(realmID, id string) error
	ListForRole(realmID, role string) ([]*Invite, error)
}
