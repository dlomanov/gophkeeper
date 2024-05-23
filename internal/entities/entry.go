package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EntryTypePassword EntryType = "password"
	EntryTypeNote     EntryType = "note"
	EntryTypeCard     EntryType = "card"
	EntryTypeBinary   EntryType = "binary"
)

type (
	EntryType string
	Entry     struct {
		ID        uuid.UUID
		Type      EntryType
		Data      []byte
		Metadata  map[string]string
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (t EntryType) Valid() bool {
	switch t {
	case EntryTypePassword, EntryTypeNote, EntryTypeCard, EntryTypeBinary:
		return true
	}
	return false
}

func NewEntry(typ EntryType, data []byte, metadata map[string]string) (*Entry, error) {
	if !typ.Valid() {
		return nil, fmt.Errorf("%w: %s", ErrEntryTypeInvalid, typ)
	}
	if data == nil {
		return nil, fmt.Errorf("%w: data empty", ErrEntryDataInvalid)
	}

	utcNow := time.Now().UTC()
	return &Entry{
		ID:        uuid.New(),
		Type:      typ,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: utcNow,
		UpdatedAt: utcNow,
	}, nil
}
