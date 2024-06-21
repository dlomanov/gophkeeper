package core

import "github.com/google/uuid"

const (
	EntryTypeUnspecified EntryType = ""
	EntryTypePassword    EntryType = "password"
	EntryTypeNote        EntryType = "note"
	EntryTypeCard        EntryType = "card"
	EntryTypeBinary      EntryType = "binary"
)

type (
	EntryType    string
	EntryVersion struct {
		ID      uuid.UUID
		Version int64
	}
)

func (t EntryType) Valid() bool {
	switch t {
	case EntryTypePassword, EntryTypeNote, EntryTypeCard, EntryTypeBinary:
		return true
	}
	return false
}
