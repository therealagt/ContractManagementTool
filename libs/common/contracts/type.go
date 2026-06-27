package contracts

type Type string

const (
	TypeNDA Type = "nda"
	TypeAVV Type = "avv"
)

func (t Type) Valid() bool {
	return t == TypeNDA || t == TypeAVV
}

func (t Type) SchemaVersion() string {
	return string(t) + "/v1"
}
