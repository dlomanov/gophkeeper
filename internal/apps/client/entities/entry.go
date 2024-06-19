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
		Entries []GetEntryResponse
	}
	GetEntryResponse struct {
		ID            uuid.UUID
		Key           string
		Type          core.EntryType
		Meta          map[string]string
		Data          EntryData
		GlobalVersion int64
		Version       int64
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}
	CreateEntryRequest struct {
		Key  string
		Type core.EntryType
		Meta map[string]string
		Data EntryData
	}
	CreateEntryResponse struct {
		ID uuid.UUID
	}
	UpdateEntryRequest struct {
		ID      uuid.UUID ``
		Meta    map[string]string
		Data    EntryData
		Version int64
	}
	DeleteEntryRequest struct {
		ID uuid.UUID
	}
	EntryData         any
	EntryDataPassword struct {
		Login    string
		Password string
	}
	EntryDataNote   string
	EntryDataBinary []byte
	EntryDataCard   struct {
		Number  string
		Expires string
		Cvc     string
		Owner   string
	}
)

func (r CreateEntryRequest) Validate() (err error) {
	if r.Key == "" {
		err = errors.Join(err, ErrEntryKeyInvalid)
	}
	if !r.Type.Valid() {
		err = errors.Join(err, ErrEntryTypeInvalid)
	}
	if !entryDataValid(r.Data) {
		err = errors.Join(err, ErrEntryTypeInvalid)
	}
	mismatchErr := func() error {
		return fmt.Errorf("%w: type and data type mismatch: %s vs %T", ErrEntryTypeInvalid, r.Type, r.Data)
	}
	switch r.Type {
	case core.EntryTypePassword:
		if _, ok := r.Data.(EntryDataPassword); !ok {
			err = errors.Join(err, mismatchErr())
		}
	case core.EntryTypeNote:
		if _, ok := r.Data.(EntryDataNote); !ok {
			err = errors.Join(err, mismatchErr())
		}
	case core.EntryTypeCard:
		if _, ok := r.Data.(EntryDataCard); !ok {
			err = errors.Join(err, mismatchErr())
		}
	case core.EntryTypeBinary:
		if _, ok := r.Data.(EntryDataBinary); !ok {
			err = errors.Join(err, mismatchErr())
		}
	default:
		err = errors.Join(err, fmt.Errorf("%w: unknown entry type: %s", ErrEntryTypeInvalid, r.Type))
	}
	return err
}

func (r UpdateEntryRequest) Validate() (err error) {
	if r.ID == uuid.Nil {
		err = errors.Join(err, ErrEntryIDInvalid)
	}
	if !entryDataValid(r.Data) {
		err = errors.Join(err, ErrEntryTypeInvalid)
	}
	return err
}

func entryDataValid(data EntryData) bool {
	switch data.(type) {
	case EntryDataPassword:
	case EntryDataNote:
	case EntryDataCard:
	case EntryDataBinary:
	default:
		return false
	}
	return true
}
