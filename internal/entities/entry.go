package entities

import (
	"fmt"
	"go.uber.org/multierr"
	"time"

	"github.com/google/uuid"
)

const (
	EntryTypeUnspecified EntryType = ""
	EntryTypePassword    EntryType = "password"
	EntryTypeNote        EntryType = "note"
	EntryTypeCard        EntryType = "card"
	EntryTypeBinary      EntryType = "binary"
)

const (
	EntryMaxDataSize = 1024 * 1024
)

type (
	Entry struct {
		ID        uuid.UUID
		UserID    uuid.UUID
		Key       string
		Type      EntryType
		Meta      map[string]string
		Data      []byte
		Version   int64
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	EntryType    string
	EntryVersion struct {
		ID      uuid.UUID
		Version int64
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

func NewEntry(
	key string,
	userID uuid.UUID,
	typ EntryType,
	data []byte,
) (*Entry, error) {
	var err error
	if key == "" {
		err = multierr.Append(err, fmt.Errorf("%w: %s", ErrEntryKeyInvalid, key))
	}
	if userID == uuid.Nil {
		err = multierr.Append(err, fmt.Errorf("%w: %s", ErrUserIDInvalid, userID))
	}
	if !typ.Valid() {
		err = multierr.Append(err, fmt.Errorf("%w: %s", ErrEntryTypeInvalid, typ))
	}
	if data == nil {
		err = multierr.Append(err, fmt.Errorf("%w: data empty", ErrEntryDataEmpty))
	}
	if len(data) > EntryMaxDataSize {
		err = fmt.Errorf("%w: data size exceeded: %d", ErrEntryDataSizeExceeded, len(data))
	}
	if err != nil {
		return nil, err
	}

	utcNow := time.Now().UTC()
	return &Entry{
		ID:        uuid.New(),
		Key:       key,
		UserID:    userID,
		Type:      typ,
		Data:      data,
		Meta:      nil,
		Version:   1,
		CreatedAt: utcNow,
		UpdatedAt: utcNow,
	}, nil
}

func (e *Entry) UpdateVersion(version int64, opts ...EntryUpdateOption) error {
	if version != e.Version {
		return fmt.Errorf("entry: version conflict: %w: %d != %d", ErrEntryVersionConflict, version, e.Version)
	}
	if err := e.update(opts...); err != nil {
		return err
	}
	e.Version++
	return nil
}

func (e *Entry) Update(opts ...EntryUpdateOption) error {
	return e.update(opts...)
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

func (e *Entry) update(opts ...EntryUpdateOption) error {
	if len(opts) == 0 {
		return nil
	}

	var err error
	for _, opt := range opts {
		err = multierr.Append(err, opt(e))
	}
	if err != nil {
		return fmt.Errorf("entry: failed to update: %w", err)
	}
	e.UpdatedAt = time.Now().UTC()
	return nil
}
