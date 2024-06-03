package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type (
	EntryUC struct {
		logger    *zap.Logger
		entryRepo EntryRepo
		encrypter Encrypter
		tx        trm.Manager
	}
	EntryRepo interface {
		Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entities.Entry, error)
		GetAll(ctx context.Context, userID uuid.UUID) ([]entities.Entry, error)
		GetByIds(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]entities.Entry, error)
		GetVersions(ctx context.Context, userID uuid.UUID) ([]entities.EntryVersion, error)
		Create(ctx context.Context, entry *entities.Entry) error
		Update(ctx context.Context, entry *entities.Entry) error
		Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	}
	Encrypter interface {
		Encrypt(data []byte) ([]byte, error)
		Decrypt(data []byte) ([]byte, error)
	}
	GetEntryRequest struct {
		UserID uuid.UUID
		ID     uuid.UUID
	}
	GetNewestEntriesRequest struct {
		UserID   uuid.UUID
		Versions map[string]int64
	}
	GetNewestEntriesResponse struct {
		Entries []entities.Entry
	}
	GetEntryResponse struct {
		Entry *entities.Entry
	}
	GetEntriesRequest struct {
		UserID uuid.UUID
	}
	GetEntriesResponse struct {
		Entries []entities.Entry
	}
	CreateEntryRequest struct {
		Key    string
		UserID uuid.UUID
		Type   entities.EntryType
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

func NewEntryUC(
	logger *zap.Logger,
	entryRepo EntryRepo,
	encrypter Encrypter,
	tx trm.Manager,
) *EntryUC {
	return &EntryUC{
		logger:    logger,
		entryRepo: entryRepo,
		encrypter: encrypter,
		tx:        tx,
	}
}

func (uc *EntryUC) Get(
	ctx context.Context,
	request GetEntryRequest,
) (response GetEntryResponse, err error) {
	if err := request.validate(); err != nil {
		return response, fmt.Errorf("get entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID

	entry, err := uc.entryRepo.Get(ctx, userID, id)
	switch {
	case errors.Is(err, entities.ErrEntryNotFound):
		uc.logger.Debug("entry not found",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return response, err
	case err != nil:
		uc.logger.Error("failed to get entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return response, err
	}
	decrypted, err := uc.encrypter.Decrypt(entry.Data)
	if err != nil {
		uc.logger.Error("failed to decrypt entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return response, fmt.Errorf("get entry: failed to decrypt entry: %w", err)
	}
	entry.Data = decrypted
	response.Entry = entry

	return response, nil
}

func (uc *EntryUC) GetEntries(
	ctx context.Context,
	request GetEntriesRequest,
) (response GetEntriesResponse, err error) {
	if err = request.validate(); err != nil {
		return response, fmt.Errorf("get all entries: invalid request: %w", err)
	}
	userID := request.UserID

	entries, err := uc.entryRepo.GetAll(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get entries",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return response, err
	}
	for i := range entries {
		decrypted, err := uc.encrypter.Decrypt(entries[i].Data)
		if err != nil {
			uc.logger.Error("failed to decrypt entry",
				zap.String("user_id", userID.String()),
				zap.Error(err))
			return response, fmt.Errorf("get all entries: failed to decrypt entry: %w", err)
		}
		entries[i].Data = decrypted
	}
	response.Entries = entries

	return response, nil
}

func (uc *EntryUC) GetNewestEntries(
	ctx context.Context,
	request GetNewestEntriesRequest,
) (response GetNewestEntriesResponse, err error) {
	if err := request.validate(); err != nil {
		return response, fmt.Errorf("get all entries: invalid request: %w", err)
	}
	userID := request.UserID

	versions, err := uc.entryRepo.GetVersions(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get versions",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return response, err
	}
	if len(versions) == 0 {
		return response, nil
	}
	var entryIds []uuid.UUID
	for _, v := range versions {
		if version, ok := request.Versions[v.ID.String()]; ok {
			if version != v.Version { // server wins
				entryIds = append(entryIds, v.ID)
			}
			continue
		}
		entryIds = append(entryIds, v.ID)
	}
	if len(entryIds) == 0 {
		return response, nil
	}

	entries, err := uc.entryRepo.GetByIds(ctx, userID, entryIds)
	if err != nil {
		uc.logger.Error("failed to get entries",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return response, err
	}
	for i := range entries {
		decrypted, err := uc.encrypter.Decrypt(entries[i].Data)
		if err != nil {
			uc.logger.Error("failed to decrypt entry",
				zap.String("user_id", userID.String()),
				zap.Error(err))
			return response, fmt.Errorf("get all entries: failed to decrypt entry: %w", err)
		}
		entries[i].Data = decrypted
	}
	response.Entries = entries

	return response, nil
}

func (uc *EntryUC) Create(
	ctx context.Context,
	request CreateEntryRequest,
) (resp CreateEntryResponse, err error) {
	if err = request.validate(); err != nil {
		return resp, fmt.Errorf("create entry: invalid request: %w", err)
	}
	userID := request.UserID

	encrypted, err := uc.encrypter.Encrypt(request.Data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return resp, fmt.Errorf("create entry: failed to encrypt entry: %w", err)
	}
	entry, err := entities.NewEntry(request.Key, userID, request.Type, encrypted)
	if err != nil {
		uc.logger.Debug("failed to create entry because of invalid arguments",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return resp, err
	}
	entry.Meta = request.Meta
	err = uc.entryRepo.Create(ctx, entry)
	switch {
	case errors.Is(err, entities.ErrEntryExists):
		uc.logger.Debug("entry already exists",
			zap.String("user_id", userID.String()),
			zap.String("entry_key", request.Key),
			zap.Error(err))
		return resp, err
	case err != nil:
		uc.logger.Error("failed to insert entry to storage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return resp, err
	}
	resp.ID = entry.ID
	resp.Version = entry.Version

	return resp, nil
}

func (uc *EntryUC) Update(
	ctx context.Context,
	request UpdateEntryRequest,
) (resp UpdateEntryResponse, err error) {
	if err = request.validate(); err != nil {
		return resp, fmt.Errorf("update entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID

	encrypted, err := uc.encrypter.Encrypt(request.Data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return resp, fmt.Errorf("update entry: failed to encrypt entry: %w", err)
	}
	var entry *entities.Entry
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		entry, err = uc.entryRepo.Get(ctx, userID, id)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			uc.logger.Debug("entry not found while updating",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		case err != nil:
			uc.logger.Error("failed to get entry while updating",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		}
		if err := entry.Update(
			request.Version,
			entities.UpdateEntryMeta(request.Meta),
			entities.UpdateEntryData(encrypted)); err != nil {
			uc.logger.Debug("failed to update entry because of invalid arguments",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		}
		if err := uc.entryRepo.Update(ctx, entry); err != nil {
			uc.logger.Error("failed to update entry in storage",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to update entry in transaction",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return resp, err
	}
	resp.ID = entry.ID
	resp.Version = entry.Version

	return resp, nil
}

func (uc *EntryUC) Delete(
	ctx context.Context,
	request DeleteEntryRequest,
) (resp DeleteEntryResponse, err error) {
	if err = request.validate(); err != nil {
		return resp, fmt.Errorf("delete entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID

	var entry *entities.Entry
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		entry, err = uc.entryRepo.Get(ctx, userID, id)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			uc.logger.Debug("entry not found while deleting",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		case err != nil:
			uc.logger.Error("failed to delete entry from storage",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		}
		err = uc.entryRepo.Delete(ctx, userID, id)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			uc.logger.Debug("entry not found while deleting",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		case err != nil:
			uc.logger.Error("failed to delete entry from storage",
				zap.String("user_id", userID.String()),
				zap.String("entry_id", id.String()),
				zap.Error(err))
			return err
		}
		return nil
	}); err != nil {
		return resp, err
	}
	resp.ID = entry.ID
	resp.Version = entry.Version

	return resp, nil
}

func (r GetEntryRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	return err
}

func (r GetNewestEntriesRequest) validate() error {
	if r.UserID == uuid.Nil {
		return entities.ErrUserIDInvalid
	}
	return nil
}

func (r GetEntriesRequest) validate() error {
	if r.UserID == uuid.Nil {
		return entities.ErrUserIDInvalid
	}
	return nil
}

func (r CreateEntryRequest) validate() error {
	var err error
	if r.Key == "" {
		err = multierr.Append(err, entities.ErrEntryKeyInvalid)
	}
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if !r.Type.Valid() {
		err = multierr.Append(err, entities.ErrEntryTypeInvalid)
	}
	if len(r.Data) == 0 {
		err = multierr.Append(err, entities.ErrEntryDataEmpty)
	}
	if len(r.Data) > entities.EntryMaxDataSize {
		err = multierr.Append(err, entities.ErrEntryDataSizeExceeded)
	}
	return err
}

func (r UpdateEntryRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	if len(r.Data) == 0 {
		err = multierr.Append(err, entities.ErrEntryDataEmpty)
	}
	if len(r.Data) > entities.EntryMaxDataSize {
		err = multierr.Append(err, entities.ErrEntryDataSizeExceeded)
	}
	if r.Version == 0 {
		err = multierr.Append(err, entities.ErrEntryVersionInvalid)
	}
	return err
}

func (r DeleteEntryRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	return err
}
