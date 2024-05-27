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
	GetRequest struct {
		UserID uuid.UUID
		ID     uuid.UUID
	}
	GetAllRequest struct {
		UserID uuid.UUID
	}
	CreateRequest struct {
		UserID uuid.UUID
		Type   entities.EntryType
		Meta   map[string]string
		Data   []byte
	}
	CreateResponse struct {
		ID        uuid.UUID
		CreatedAt time.Time
	}
	UpdateRequest struct {
		ID     uuid.UUID
		UserID uuid.UUID
		Type   entities.EntryType
		Meta   map[string]string
		Data   []byte
	}
	UpdateResponse struct {
		ID        uuid.UUID
		UpdatedAt time.Time
	}
	DeleteRequest struct {
		ID     uuid.UUID
		UserID uuid.UUID
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
	request GetRequest,
) (*entities.Entry, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("get entry: invalid request: %w", err)
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
		return nil, err
	case err != nil:
		uc.logger.Error("failed to get entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return nil, err
	}
	return entry, nil

}

func (uc *EntryUC) GetAll(
	ctx context.Context,
	request GetAllRequest,
) ([]entities.Entry, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("get all entries: invalid request: %w", err)
	}
	userID := request.UserID

	entries, err := uc.entryRepo.GetAll(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get entries",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}
	return entries, nil
}

func (uc *EntryUC) Create(
	ctx context.Context,
	request CreateRequest,
) (*CreateResponse, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("create entry: invalid request: %w", err)
	}
	userID := request.UserID

	entry, err := entities.NewEntry(userID, request.Type, request.Data)
	if err != nil {
		uc.logger.Debug("failed to create entry because of invalid arguments",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}
	entry.Meta = request.Meta
	if err := uc.entryRepo.Create(ctx, entry); err != nil {
		uc.logger.Error("failed to insert entry to storage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}
	return &CreateResponse{
		ID:        entry.ID,
		CreatedAt: entry.CreatedAt,
	}, nil
}

func (uc *EntryUC) Update(
	ctx context.Context,
	request UpdateRequest,
) (*UpdateResponse, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("update entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID

	var entry *entities.Entry
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
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
			entities.UpdateEntryType(request.Type),
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
	})
	if err != nil {
		uc.logger.Error("failed to update entry in storage",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return nil, err
	}

	return &UpdateResponse{
		ID:        entry.ID,
		UpdatedAt: entry.UpdatedAt,
	}, nil

}

func (uc *EntryUC) Delete(
	ctx context.Context,
	request DeleteRequest,
) error {
	if err := request.validate(); err != nil {
		return fmt.Errorf("delete entry: invalid request: %w", err)
	}
	userID := request.UserID
	id := request.ID

	err := uc.entryRepo.Delete(ctx, userID, id)
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
}

func (r GetRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	return err
}

func (r GetAllRequest) validate() error {
	if r.UserID == uuid.Nil {
		return entities.ErrUserIDInvalid
	}
	return nil
}

func (r CreateRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if !r.Type.Valid() {
		err = multierr.Append(err, entities.ErrEntryTypeInvalid)
	}
	if r.Data == nil {
		err = multierr.Append(err, entities.ErrEntryDataEmpty)
	}
	return err
}

func (r UpdateRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	if !r.Type.Valid() {
		err = multierr.Append(err, entities.ErrEntryTypeInvalid)
	}
	if r.Data == nil {
		err = multierr.Append(err, entities.ErrEntryDataEmpty)
	}
	return err
}

func (r DeleteRequest) validate() error {
	var err error
	if r.UserID == uuid.Nil {
		err = multierr.Append(err, entities.ErrUserIDInvalid)
	}
	if r.ID == uuid.Nil {
		err = multierr.Append(err, entities.ErrEntryIDInvalid)
	}
	return err
}
