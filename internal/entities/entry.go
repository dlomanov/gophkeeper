package entities

import (
	"fmt"
	"go.uber.org/multierr"
	"time"

	"github.com/google/uuid"
)

const (
	EntryTypePassword EntryType = "password"
	EntryTypeNote     EntryType = "note"
	EntryTypeCard     EntryType = "card"
	EntryTypeBinary   EntryType = "binary"
)

const (
	EntryMaxDataSize = 1024 * 1024
)

type (
	EntryType string
	Entry     struct {
		ID        uuid.UUID
		UserID    uuid.UUID
		Type      EntryType
		Meta      map[string]string
		Data      []byte
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	EntryUpdateOption func(e *Entry) error
)

func (t EntryType) Valid() bool {
	switch t {
	case EntryTypePassword, EntryTypeNote, EntryTypeCard, EntryTypeBinary:
		return true
	}
	return false
}

func NewEntry(userID uuid.UUID, typ EntryType, data []byte) (*Entry, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: %s", ErrUserIDInvalid, userID)
	}
	if !typ.Valid() {
		return nil, fmt.Errorf("%w: %s", ErrEntryTypeInvalid, typ)
	}
	if data == nil {
		return nil, fmt.Errorf("%w: data empty", ErrEntryDataEmpty)
	}

	if len(data) > EntryMaxDataSize {
		return nil, fmt.Errorf("%w: data size exceeded: %d", ErrEntryDataSizeExceeded, len(data))
	}

	utcNow := time.Now().UTC()
	return &Entry{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      typ,
		Data:      data,
		Meta:      nil,
		CreatedAt: utcNow,
		UpdatedAt: utcNow,
	}, nil
}

func (e *Entry) Update(opts ...EntryUpdateOption) error {
	if len(opts) == 0 {
		return nil
	}

	var err error
	for _, opt := range opts {
		err = multierr.Append(err, opt(e))
	}
	if err != nil {
		return err
	}
	e.UpdatedAt = time.Now().UTC()
	return nil
}

func UpdateEntryType(typ EntryType) EntryUpdateOption {
	return func(e *Entry) error {
		if !typ.Valid() {
			return fmt.Errorf("%w: %s", ErrEntryTypeInvalid, typ)
		}
		e.Type = typ
		return nil
	}
}

func UpdateEntryData(data []byte) EntryUpdateOption {
	return func(e *Entry) error {
		if len(data) == 0 {
			return fmt.Errorf("%w: data empty", ErrEntryDataEmpty)
		}
		if len(data) > EntryMaxDataSize {
			return fmt.Errorf("%w: data size exceeded: %d", ErrEntryDataSizeExceeded, len(data))
		}
		e.Data = data
		return nil
	}
}

func UpdateEntryMeta(meta map[string]string) EntryUpdateOption {
	return func(e *Entry) error {
		e.Meta = meta
		return nil
	}
}
