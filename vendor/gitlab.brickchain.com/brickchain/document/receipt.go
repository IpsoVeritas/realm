package document

import "time"

const ReceiptType = "receipt"

type Interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Receipt struct {
	Base
	Role      string            `json:"role,omitempty"`
	Action    string            `json:"action,omitempty"`
	URI       string            `json:"viewuri,omitempty"`
	JWT       string            `json:"jwt,omitempty"`
	Intervals []Interval        `json:"intervals,omitempty"`
	Label     string            `json:"label,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

func NewReceipt(role string) *Receipt {
	return &Receipt{
		Base: Base{
			Context:   Context,
			Type:      ReceiptType,
			Timestamp: time.Now(),
		},
		Role: role,
	}
}
