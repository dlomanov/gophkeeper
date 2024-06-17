package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
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
		GetVersions(ctx context.Context, userID uuid.UUID) ([]core.EntryVersion, error)
		Create(ctx context.Context, entry *entities.Entry) error
		Update(ctx context.Context, entry *entities.Entry) error
		Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	}
	EntryDiffer interface {
		GetDiff(
			ctx context.Context,
			serverVersions []core.EntryVersion,
			clientVersions []core.EntryVersion,
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
	request entities.GetEntryRequest,
) (response entities.GetEntryResponse, err error) {
	if err := request.Validate(); err != nil {
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
	request entities.GetEntriesRequest,
) (response entities.GetEntriesResponse, err error) {
	if err = request.Validate(); err != nil {
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
	request entities.GetEntriesDiffRequest,
) (response entities.GetEntriesDiffResponse, err error) {
	if err := request.Validate(); err != nil {
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
	request entities.CreateEntryRequest,
) (response entities.CreateEntryResponse, err error) {
	if err = request.Validate(); err != nil {
		return response, fmt.Errorf("create entry: invalid request: %w", err)
	}
	userID := request.UserID

	encrypted, err := uc.encrypter.Encrypt(request.Data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return response, fmt.Errorf("create_entry: failed to encrypt entry: %w", err)
	}
	entry, err := entities.NewEntry(request.Key, userID, request.Type, encrypted)
	if err != nil {
		return response, fmt.Errorf("create_entry: failed to create_entry: %w", err)
	}
	entry.Meta = request.Meta
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		err = uc.entryRepo.Create(ctx, entry)
		switch {
		case errors.Is(err, entities.ErrEntryExists):
			conflictKey := uc.newConflictKey(entry.Key, entry.Version)
			conflictEntry, err := entities.NewEntry(conflictKey, entry.UserID, entry.Type, encrypted)
			if err != nil {
				return fmt.Errorf("create_entry: failed to create conflict entry: %w: %w", err, entities.ErrEntryExists)
			}
			conflictEntry.Meta = request.Meta
			if err = uc.entryRepo.Create(ctx, conflictEntry); err != nil {
				return fmt.Errorf("create_entry: failed to create conflict entry in repo: %w: %w", err, entities.ErrEntryExists)
			}
			entry = conflictEntry
		case err != nil:
			return fmt.Errorf("create_entry: failed to create entry in repo: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to create entry",
			zap.String("user_id", userID.String()),
			zap.String("key", request.Key),
			zap.Error(err))
		return response, err
	}
	response.ID = entry.ID
	response.Version = entry.Version

	return response, nil
}

func (uc *EntryUC) Update(
	ctx context.Context,
	request entities.UpdateEntryRequest,
) (response entities.UpdateEntryResponse, err error) {
	if err = request.Validate(); err != nil {
		return response, fmt.Errorf("update entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID
	version := request.Version

	encrypted, err := uc.encrypter.Encrypt(request.Data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return response, fmt.Errorf("update_entry: failed to encrypt entry: %w", err)
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
		err = entry.Update(
			version,
			entities.UpdateEntryMeta(request.Meta),
			entities.UpdateEntryData(encrypted))
		switch {
		// handle version conflict by saving conflict version of entry
		case errors.Is(err, entities.ErrEntryVersionConflict):
			conflictKey := uc.newConflictKey(entry.Key, request.Version)
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
		return response, err
	}
	response.ID = entry.ID
	response.Version = entry.Version

	return response, nil
}

func (uc *EntryUC) Delete(
	ctx context.Context,
	request entities.DeleteEntryRequest,
) (response entities.DeleteEntryResponse, err error) {
	if err = request.Validate(); err != nil {
		return response, fmt.Errorf("delete entry: invalid request: %w", err)
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
		return response, err
	}
	response.ID = entry.ID
	response.Version = entry.Version

	return response, nil
}

func (uc *EntryUC) newConflictKey(key string, version int64) string {
	return fmt.Sprintf("%s_conflict_%d_%s", key, version, uuid.New().String())
}
