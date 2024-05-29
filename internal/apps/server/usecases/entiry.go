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
	"time"
)

type (
	EntryUC struct {
		logger    *zap.Logger
		entryRepo EntryRepo
		tx        trm.Manager
	}
	EntryRepo interface {
		Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entities.Entry, error)
		GetAll(ctx context.Context, userID uuid.UUID) ([]entities.Entry, error)
		Create(ctx context.Context, entry *entities.Entry) error
		Update(ctx context.Context, entry *entities.Entry) error
		Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	}
	GetEntryRequest struct {
		UserID uuid.UUID
		ID     uuid.UUID
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
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	UpdateEntryRequest struct {
		ID     uuid.UUID
		UserID uuid.UUID
		Meta   map[string]string
		Data   []byte
	}
	UpdateEntryResponse struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	DeleteEntryRequest struct {
		ID     uuid.UUID
		UserID uuid.UUID
	}
	DeleteEntryResponse struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func NewEntryUC(
	logger *zap.Logger,
	entryRepo EntryRepo,
	tx trm.Manager,
) *EntryUC {
	return &EntryUC{
		logger:    logger,
		entryRepo: entryRepo,
		tx:        tx,
	}
}

func (uc *EntryUC) Get(
	ctx context.Context,
	request GetEntryRequest,
) (GetEntryResponse, error) {
	response := GetEntryResponse{}

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
	response.Entry = entry

	return response, nil

}

func (uc *EntryUC) GetAll(
	ctx context.Context,
	request GetEntriesRequest,
) (GetEntriesResponse, error) {
	response := GetEntriesResponse{}
	if err := request.validate(); err != nil {
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

	entry, err := entities.NewEntry(request.Key, userID, request.Type, request.Data)
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
	resp.CreatedAt = entry.CreatedAt
	resp.UpdatedAt = entry.UpdatedAt

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
			entities.UpdateEntryMeta(request.Meta),
			entities.UpdateEntryData(request.Data)); err != nil {
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
		uc.logger.Error("failed to update entry in storage",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return resp, err
	}
	resp.ID = entry.ID
	resp.CreatedAt = entry.CreatedAt
	resp.UpdatedAt = entry.UpdatedAt

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
	resp.CreatedAt = entry.CreatedAt
	resp.UpdatedAt = entry.UpdatedAt

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
