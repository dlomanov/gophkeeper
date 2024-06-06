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
		logger      *zap.Logger
		entryRepo   EntryRepo
		entryDiffer EntryDiffer
		encrypter   Encrypter
		tx          trm.Manager
	}
	EntryRepo interface {
		Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entities.Entry, error)
		GetAll(ctx context.Context, userID uuid.UUID) ([]entities.Entry, error)
		GetByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]entities.Entry, error)
		GetVersions(ctx context.Context, userID uuid.UUID) ([]entities.EntryVersion, error)
		Create(ctx context.Context, entry *entities.Entry) error
		Update(ctx context.Context, entry *entities.Entry) error
		Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	}
	EntryDiffer interface {
		GetDiff(
			ctx context.Context,
			serverVersions []entities.EntryVersion,
			clientVersions []entities.EntryVersion,
		) (
			createIDs []uuid.UUID,
			updateIDs []uuid.UUID,
			deleteIDs []uuid.UUID,
			err error,
		)
	}
	Encrypter interface {
		Encrypt(data []byte) ([]byte, error)
		Decrypt(data []byte) ([]byte, error)
	}
	GetEntryRequest struct {
		UserID uuid.UUID
		ID     uuid.UUID
	}
	GetEntriesDiffRequest struct {
		UserID   uuid.UUID
		Versions []entities.EntryVersion
	}
	GetEntriesDiffResponse struct {
		Entries   []entities.Entry
		CreateIDs []uuid.UUID
		UpdateIDs []uuid.UUID
		DeleteIDs []uuid.UUID
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
	entryDiffer EntryDiffer,
	encrypter Encrypter,
	tx trm.Manager,
) *EntryUC {
	return &EntryUC{
		logger:      logger,
		entryRepo:   entryRepo,
		entryDiffer: entryDiffer,
		encrypter:   encrypter,
		tx:          tx,
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

func (uc *EntryUC) GetEntriesDiff(
	ctx context.Context,
	request GetEntriesDiffRequest,
) (response GetEntriesDiffResponse, err error) {
	if err := request.validate(); err != nil {
		return response, fmt.Errorf("get all entries: invalid request: %w", err)
	}
	var (
		userID    = request.UserID
		createIDs []uuid.UUID
		updateIDs []uuid.UUID
		deleteIDs []uuid.UUID
		entries   []entities.Entry
	)
	if err := uc.tx.Do(ctx, func(ctx context.Context) error {
		versions, err := uc.entryRepo.GetVersions(ctx, userID)
		if err != nil {
			return fmt.Errorf("get_entries_diff: failed to get versions from storage: %w", err)
		}
		createIDs, updateIDs, deleteIDs, err = uc.entryDiffer.GetDiff(ctx, versions, request.Versions)
		if err != nil {
			return fmt.Errorf("get_entries_diff: failed to diff versions: %w", err)
		}
		entryIDs := make([]uuid.UUID, 0, len(createIDs)+len(updateIDs))
		entryIDs = append(entryIDs, createIDs...)
		entryIDs = append(entryIDs, updateIDs...)
		entries, err = uc.entryRepo.GetByIDs(ctx, userID, entryIDs)
		if err != nil {
			return fmt.Errorf("get_entries_diff: failed to get entries from storage: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to calculate diff",
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
	response.CreateIDs = createIDs
	response.UpdateIDs = updateIDs
	response.DeleteIDs = deleteIDs
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
		return resp, fmt.Errorf("create_entry: failed to encrypt entry: %w", err)
	}
	entry, err := entities.NewEntry(request.Key, userID, request.Type, encrypted)
	if err != nil {
		return resp, fmt.Errorf("create_entry: failed to create_entry: %w", err)
	}
	entry.Meta = request.Meta
	err = uc.entryRepo.Create(ctx, entry)
	switch {
	case errors.Is(err, entities.ErrEntryExists):
		return resp, fmt.Errorf("create_entry: entry already exists: %w", err)
	case err != nil:
		return resp, fmt.Errorf("create_entry: failed to insert entry to storage: %w", err)
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
	version := request.Version

	encrypted, err := uc.encrypter.Encrypt(request.Data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return resp, fmt.Errorf("update_entry: failed to encrypt entry: %w", err)
	}
	var entry *entities.Entry
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		entry, err = uc.entryRepo.Get(ctx, userID, id)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			return fmt.Errorf("update_entry: entry not found: %w", entities.ErrEntryNotFound)
		case err != nil:
			return fmt.Errorf("update_entry: failed to get entry from storage: %w", err)
		}
		err = entry.UpdateVersion(
			version,
			entities.UpdateEntryMeta(request.Meta),
			entities.UpdateEntryData(encrypted))
		switch {
		// handle version conflict by saving conflict version of entry
		case errors.Is(err, entities.ErrEntryVersionConflict):
			conflictKey := fmt.Sprintf("%s_conflict_%d_%s", entry.Key, request.Version, uuid.New().String())
			conflictEntry, err := entities.NewEntry(conflictKey, entry.UserID, entry.Type, encrypted)
			if err != nil {
				return fmt.Errorf("update_entry: failed to create conflict entry: %w", err)
			}
			conflictEntry.Meta = request.Meta
			if err = uc.entryRepo.Create(ctx, conflictEntry); err != nil {
				return fmt.Errorf("update_entry: failed to create conflict entry in storage: %w", err)
			}
			entry = conflictEntry
			return nil
		case errors.Is(err, entities.ErrEntryVersionInvalid):
			return fmt.Errorf("update_entry: invalid entry version: %w", err)
		case err != nil:
			return fmt.Errorf("update_entry: failed to update entry: %w", err)
		}
		if err := uc.entryRepo.Update(ctx, entry); err != nil {
			return fmt.Errorf("update_entry: failed to update entry in storage: %w", err)
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
			return fmt.Errorf("delete_entry: entry not found: %w", entities.ErrEntryNotFound)
		case err != nil:
			return fmt.Errorf("delete_entry: failed to get entry from storage: %w", err)
		}
		err = uc.entryRepo.Delete(ctx, userID, id)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			return fmt.Errorf("delete_entry: entry not found: %w", entities.ErrEntryNotFound)
		case err != nil:
			return fmt.Errorf("delete_entry: failed to delete entry from storage: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to delete entry from storage",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
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

func (r GetEntriesRequest) validate() error {
	if r.UserID == uuid.Nil {
		return entities.ErrUserIDInvalid
	}
	return nil
}

func (r GetEntriesDiffRequest) validate() error {
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
