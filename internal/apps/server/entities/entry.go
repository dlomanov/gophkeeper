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
		ID        uuid.UUID
		UserID    uuid.UUID
		Key       string
		Type      core.EntryType
		Meta      map[string]string
		Data      []byte
		Version   int64
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	EntryUpdateOption func(e *Entry) error
)

func NewEntry(
	key string,
	userID uuid.UUID,
	typ core.EntryType,
	data []byte,
) (*Entry, error) {
	var err error
	if key == "" {
		err = errors.Join(err, fmt.Errorf("%w: %s", ErrEntryKeyInvalid, key))
	}
	if userID == uuid.Nil {
		err = errors.Join(err, fmt.Errorf("%w: %s", ErrUserIDInvalid, userID))
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

func (e *Entry) Update(version int64, opts ...EntryUpdateOption) error {
	if version != e.Version {
		return fmt.Errorf("entry: version conflict: %w: %d != %d", ErrEntryVersionConflict, version, e.Version)
	}
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

type (
	GetEntryRequest struct {
		UserID uuid.UUID
		ID     uuid.UUID
	}
	GetEntriesDiffRequest struct {
		UserID   uuid.UUID
		Versions []core.EntryVersion
	}
	GetEntriesDiffResponse struct {
		Entries   []Entry
		CreateIDs []uuid.UUID
		UpdateIDs []uuid.UUID
		DeleteIDs []uuid.UUID
	}
	GetEntryResponse struct {
		Entry *Entry
	}
	GetEntriesRequest struct {
		UserID uuid.UUID
	}
	GetEntriesResponse struct {
		Entries []Entry
	}
	CreateEntryRequest struct {
		Key    string
		UserID uuid.UUID
		Type   core.EntryType
		Meta   map[string]string
		Data   []byte
	}
	CreateEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
	UpdateEntryRequest struct {
		ID      uuid.UUID
		UserID  uuid.UUID
		Meta    map[string]string
		Data    []byte
		Version int64
	}
	UpdateEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
	DeleteEntryRequest struct {
		ID     uuid.UUID
		UserID uuid.UUID
	}
	DeleteEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
)

func (r GetEntryRequest) Validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = errors.Join(err, ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = errors.Join(err, ErrEntryIDInvalid)
	}
	return err
}

func (r GetEntriesRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return ErrUserIDInvalid
	}
	return nil
}

func (r GetEntriesDiffRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return ErrUserIDInvalid
	}
	return nil
}

func (r CreateEntryRequest) Validate() error {
	var err error
	if r.Key == "" {
		err = errors.Join(err, ErrEntryKeyInvalid)
	}
	if r.UserID == uuid.Nil {
		err = errors.Join(err, ErrUserIDInvalid)
	}
	if !r.Type.Valid() {
		err = errors.Join(err, ErrEntryTypeInvalid)
	}
	if len(r.Data) == 0 {
		err = errors.Join(err, ErrEntryDataEmpty)
	}
	if len(r.Data) > EntryMaxDataSize {
		err = errors.Join(err, ErrEntryDataSizeExceeded)
	}
	return err
}

func (r UpdateEntryRequest) Validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = errors.Join(err, ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = errors.Join(err, ErrEntryIDInvalid)
	}
	if len(r.Data) == 0 {
		err = errors.Join(err, ErrEntryDataEmpty)
	}
	if len(r.Data) > EntryMaxDataSize {
		err = errors.Join(err, ErrEntryDataSizeExceeded)
	}
	if r.Version == 0 {
		err = errors.Join(err, ErrEntryVersionInvalid)
	}
	return err
}

func (r DeleteEntryRequest) Validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = errors.Join(err, ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = errors.Join(err, ErrEntryIDInvalid)
	}
	return err
}
