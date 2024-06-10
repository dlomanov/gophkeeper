package entities

import (
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/core"
	"go.uber.org/multierr"
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
	EntryUpdateOption func(e *Entry) error
)

func NewEntry(
	key string,
	typ core.EntryType,
	data []byte,
) (*Entry, error) {
	var err error
	if key == "" {
		err = multierr.Append(err, fmt.Errorf("%w: %s", ErrEntryKeyInvalid, key))
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
		err = multierr.Append(err, opt(e))
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