package document

import "encoding/json"

type BaseInterface interface {
	GetCertificateChain() string
	Expand() string
	GetType() string
	GetSubType() string
}

func Serialize(doc BaseInterface) (string, string) {
	docBytes, _ := json.Marshal(doc)
	return string(docBytes), "application/json"
}

func Deserialize(data string, doc interface{}) error {
	return json.Unmarshal([]byte(data), doc)
}

func GetType(data string) string {
	b := Base{}
	err := Deserialize(data, &b)
	if err != nil || b.Expand() == "" {
		return "unknown"
	}
	return b.Expand()
}

func GetSubType(data string) string {
	b := Base{}
	err := Deserialize(data, &b)
	if err != nil || b.Expand() == "" {
		return "unknown"
	}
	return b.SubType
}
