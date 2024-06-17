package entities

import (
	"errors"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/core"
	"time"

	"github.com/google/uuid"
)

const (
	EntryMaxDataSize = 1024 * 1024
)

type (
	Entry struct {
		ID            uuid.UUID
		Key           string
		Type          core.EntryType
		Meta          map[string]string
		Data          []byte
		GlobalVersion int64
		Version       int64
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}
	EntrySync struct {
		ID        uuid.UUID
		CreatedAt time.Time
	}
	EntryUpdateOption func(e *Entry) error
)

func NewEntry(
	key string,
	typ core.EntryType,
	data []byte,
) (*Entry, error) {
	var err error
	if key == "" {
		err = errors.Join(err, fmt.Errorf("%w: %s", ErrEntryKeyInvalid, key))
	}
	if !typ.Valid() {
		err = errors.Join(err, fmt.Errorf("%w: %s", ErrEntryTypeInvalid, typ))
	}
	if data == nil {
		err = errors.Join(err, fmt.Errorf("%w: data empty", ErrEntryDataEmpty))
	}
	if len(data) > EntryMaxDataSize {
		err = fmt.Errorf("%w: data size exceeded: %d", ErrEntryDataSizeExceeded, len(data))
	}
	if err != nil {
		return nil, err
	}

	utcNow := time.Now().UTC()
	return &Entry{
		ID:            uuid.New(),
		Key:           key,
		Type:          typ,
		Data:          data,
		Meta:          nil,
		Version:       1,
		GlobalVersion: 0,
		CreatedAt:     utcNow,
		UpdatedAt:     utcNow,
	}, nil
}

func (e *Entry) Update(opts ...EntryUpdateOption) error {
	if len(opts) == 0 {
		return nil
	}

	var err error
	for _, opt := range opts {
		err = errors.Join(err, opt(e))
	}
	if err != nil {
		return fmt.Errorf("entry: failed to update: %w", err)
	}
	e.UpdatedAt = time.Now().UTC()
	e.Version++
	return nil
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

func NewEntrySync(id uuid.UUID) *EntrySync {
	return &EntrySync{
		ID:        id,
		CreatedAt: time.Now().UTC(),
	}
}

type (
	GetEntriesResponse struct {
		Entries []Entry
	}
	CreateEntryRequest struct {
		Key  string
		Type core.EntryType
		Meta map[string]string
		Data []byte
	}
	CreateEntryResponse struct {
		ID uuid.UUID
	}
	UpdateEntryRequest struct {
		ID      uuid.UUID ``
		Meta    map[string]string
		Data    []byte
		Version int64
	}
	DeleteEntryRequest struct {
		ID uuid.UUID
	}
)
