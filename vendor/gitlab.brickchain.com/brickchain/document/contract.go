package document

import "time"

const ContractType = "contract"

type Contract struct {
	Base
	Text string `json:"text,omitempty"`
}

func NewContract() *Contract {
	return &Contract{
		Base: Base{
			Context:   Context,
			Type:      ContractType,
			Timestamp: time.Now(),
		},
	}
}
