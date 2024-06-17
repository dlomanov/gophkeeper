package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/shared/mapper"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

type (
	EntryUC struct {
		logger        *zap.Logger
		cache         *mem.Cache
		entryClient   pb.EntryServiceClient
		entryRepo     EntryRepo
		entrySyncRepo EntrySyncRepo
		encrypter     Encrypter
		mapper        mapper.EntryMapper
		tx            trm.Manager
	}
	EntryRepo interface {
		Get(ctx context.Context, id uuid.UUID) (entities.Entry, error)
		GetAll(ctx context.Context) ([]entities.Entry, error)
		GetVersions(ctx context.Context) ([]core.EntryVersion, error)
		Delete(ctx context.Context, id uuid.UUID) error
		Create(ctx context.Context, entry entities.Entry) error
		Update(ctx context.Context, entry entities.Entry) error
	}
	EntrySyncRepo interface {
		GetAll(ctx context.Context) ([]entities.EntrySync, error)
		Delete(ctx context.Context, id uuid.UUID) error
		Create(ctx context.Context, entrySync entities.EntrySync) error
	}
	Encrypter interface {
		Encrypt(data []byte) ([]byte, error)
		Decrypt(data []byte) ([]byte, error)
	}
)

func NewEntriesUC(
	logger *zap.Logger,
	entryClient pb.EntryServiceClient,
	entryRepo EntryRepo,
	entrySyncRepo EntrySyncRepo,
	encrypter Encrypter,
	cache *mem.Cache,
	tx trm.Manager,
) *EntryUC {
	return &EntryUC{
		logger:        logger,
		cache:         cache,
		entryClient:   entryClient,
		entryRepo:     entryRepo,
		entrySyncRepo: entrySyncRepo,
		encrypter:     encrypter,
		mapper:        mapper.EntryMapper{},
		tx:            tx,
	}
}

func (uc *EntryUC) GetAll(ctx context.Context) (response entities.GetEntriesResponse, err error) {
	entries, err := uc.entryRepo.GetAll(ctx)
	if err != nil {
		return response, fmt.Errorf("entry_usecase: failed to get entries: %w", err)
	}
	for i, v := range entries {
		data, err := uc.encrypter.Decrypt(v.Data)
		if err != nil {
			return response, fmt.Errorf("entry_usecase: failed to decrypt entry: %w", err)
		}
		entries[i].Data = data
	}
	response.Entries = entries
	return response, err
}

func (uc *EntryUC) Create(
	ctx context.Context,
	request entities.CreateEntryRequest,
) (response entities.CreateEntryResponse, err error) {
	var data []byte
	if data, err = uc.encrypter.Encrypt(request.Data); err != nil {
		uc.logger.Error("failed to encrypt entry data", zap.Error(err))
		return response, fmt.Errorf("entry_usecase: failed to encrypt entry data: %w", err)
	}
	entry, err := entities.NewEntry(request.Key, request.Type, data)
	if err != nil {
		uc.logger.Error("failed to encrypt entry data", zap.Error(err))
		return response, fmt.Errorf("entry_usecase: %w", err)
	}
	entry.Meta = request.Meta
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		err = uc.entryRepo.Create(ctx, *entry)
		switch {
		case errors.Is(err, entities.ErrEntryExists):
			return fmt.Errorf("entry_usecase: %w", err)
		case err != nil:
			return fmt.Errorf("entry_usecase: failed to create entry in repo: %w", err)
		}
		if err = uc.entrySyncRepo.Create(ctx, *entities.NewEntrySync(entry.ID)); err != nil {
			return fmt.Errorf("entry_usecase: failed to create entry sync in repo: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to create entry", zap.Error(err))
		return response, err
	}
	response.ID = entry.ID
	return response, nil
}

func (uc *EntryUC) Update(
	ctx context.Context,
	request entities.UpdateEntryRequest,
) (err error) {
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		entry, err := uc.entryRepo.Get(ctx, request.ID)
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			return fmt.Errorf("entry_usecase: %w", err)
		case err != nil:
			return fmt.Errorf("entry_usecase: failed to get entry: %w", err)
		}
		data, err := uc.encrypter.Encrypt(request.Data)
		if err != nil {
			return fmt.Errorf("entry_usecase: failed to encrypt entry data: %w", err)
		}
		if err = entry.Update(
			entities.UpdateEntryMeta(request.Meta),
			entities.UpdateEntryData(data)); err != nil {
			return fmt.Errorf("entry_usecase: %w", err)
		}
		if err = uc.entryRepo.Update(ctx, entry); err != nil {
			return fmt.Errorf("entry_usecase: failed to update entry in repo: %w", err)
		}
		if err = uc.entrySyncRepo.Create(ctx, entities.EntrySync{
			ID:        entry.ID,
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			return fmt.Errorf("entry_usecase: failed to create entry sync in repo: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to update entry", zap.Error(err))
		return err
	}
	return nil
}

func (uc *EntryUC) Delete(
	ctx context.Context,
	request entities.DeleteEntryRequest,
) (err error) {
	if err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err = uc.entryRepo.Delete(ctx, request.ID); err != nil {
			return fmt.Errorf("entry_usecase: failed to delete entry in repo: %w", err)
		}
		if err = uc.entrySyncRepo.Create(ctx, *entities.NewEntrySync(request.ID)); err != nil {
			return fmt.Errorf("entry_usecase: failed to create entry sync in repo: %w", err)
		}
		return nil
	}); err != nil {
		uc.logger.Error("failed to delete entry", zap.Error(err))
		return err
	}
	return nil
}

func (uc *EntryUC) Sync(ctx context.Context) (err error) {
	if ctx, err = uc.appendToken(ctx); err != nil {
		return err
	}
	if err = uc.pushEntries(ctx); err != nil {
		uc.logger.Error("failed to push changes", zap.Error(err))
		return fmt.Errorf("entry_usecase: failed to push changes: %w", err)
	}
	if err = uc.fetch(ctx); err != nil {
		uc.logger.Error("failed to fetch changes", zap.Error(err))
		return fmt.Errorf("entry_usecase: failed to fetch changes: %w", err)
	}
	return nil
}

func (uc *EntryUC) pushEntries(ctx context.Context) error {
	result, err := uc.entrySyncRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("entry_usecase: failed to get entry syncs: %w", err)
	}
	for _, v := range result {
		if err = uc.pushEntry(ctx, v.ID); err != nil {
			uc.logger.Error("failed to push entry", zap.Error(err))
			continue
		}
		if err = uc.entrySyncRepo.Delete(ctx, v.ID); err != nil {
			uc.logger.Error("failed to delete entry sync", zap.Error(err))
		}
	}
	return nil
}

func (uc *EntryUC) pushEntry(ctx context.Context, id uuid.UUID) error {
	type pushType int
	const (
		pushTypeNone pushType = iota
		pushTypeCreate
		pushTypeUpdate
		pushTypeDelete
	)

	getPushType := func(entry entities.Entry, err error) pushType {
		switch {
		case errors.Is(err, entities.ErrEntryNotFound):
			return pushTypeDelete
		case err != nil:
			return pushTypeNone
		case entry.GlobalVersion == 0:
			return pushTypeCreate
		default:
			return pushTypeUpdate
		}
	}

	entry, err := uc.entryRepo.Get(ctx, id)
	typ := getPushType(entry, err)
	switch typ {
	case pushTypeCreate:
		data, err := uc.encrypter.Decrypt(entry.Data)
		if err != nil {
			return fmt.Errorf("entry_usecase: failed to decrypt entry data: %w", err)
		}
		_, err = uc.entryClient.Create(ctx, &pb.CreateEntryRequest{
			Key:  entry.Key,
			Type: uc.mapper.ToAPIType(entry.Type),
			Meta: entry.Meta,
			Data: data,
		})
		switch {
		case status.Code(err) == codes.InvalidArgument:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrEntryInvalid, err)
		case status.Code(err) == codes.AlreadyExists:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrEntryExists, err)
		case status.Code(err) == codes.Unavailable:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrServerUnavailable, err)
		case status.Code(err) == codes.Unauthenticated:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrUserTokenInvalid, err)
		case err != nil:
			return fmt.Errorf("entry_usecase: failed to create entry: %w", err)
		}
	case pushTypeUpdate:
		data, err := uc.encrypter.Decrypt(entry.Data)
		if err != nil {
			return fmt.Errorf("entry_usecase: failed to decrypt entry data: %w", err)
		}
		_, err = uc.entryClient.Update(ctx, &pb.UpdateEntryRequest{
			Id:      id.String(),
			Meta:    entry.Meta,
			Data:    data,
			Version: entry.GlobalVersion,
		})
		switch {
		case status.Code(err) == codes.InvalidArgument:
			return fmt.Errorf("entry_usecase: failed to update entry: %w: %w", entities.ErrEntryInvalid, err)
		case status.Code(err) == codes.NotFound:
			return fmt.Errorf("entry_usecase: failed to update entry: %w: %w", entities.ErrEntryNotFound, err)
		case status.Code(err) == codes.Unavailable:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrServerUnavailable, err)
		case status.Code(err) == codes.Unauthenticated:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrUserTokenInvalid, err)
		case err != nil:
			return fmt.Errorf("entry_usecase: failed to update entry: %w", err)
		}
	case pushTypeDelete:
		_, err = uc.entryClient.Delete(ctx, &pb.DeleteEntryRequest{Id: id.String()})
		switch {
		case status.Code(err) == codes.InvalidArgument:
			return fmt.Errorf("entry_usecase: failed to delete entry: %w: %w", entities.ErrEntryInvalid, err)
		case status.Code(err) == codes.NotFound:
			return fmt.Errorf("entry_usecase: failed to delete entry: %w: %w", entities.ErrEntryNotFound, err)
		case status.Code(err) == codes.Unavailable:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrServerUnavailable, err)
		case status.Code(err) == codes.Unauthenticated:
			return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrUserTokenInvalid, err)
		case err != nil:
			return fmt.Errorf("entry_usecase: failed to delete entry: %w", err)
		}
	default:
		return fmt.Errorf("entry_usecase: failed to get entry: %w", err)
	}
	return nil
}

func (uc *EntryUC) fetch(ctx context.Context) error {
	entries, err := uc.entryRepo.GetVersions(ctx)
	if err != nil {
		return fmt.Errorf("entry_usecase: failed to get entry versions: %w", err)
	}
	versions := make([]*pb.EntryVersion, len(entries))
	for i, v := range entries {
		versions[i] = &pb.EntryVersion{
			Id:      v.ID.String(),
			Version: v.Version,
		}
	}
	resp, err := uc.entryClient.GetDiff(ctx, &pb.GetEntriesDiffRequest{Versions: versions})
	switch {
	case status.Code(err) == codes.Unavailable:
		return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrServerUnavailable, err)
	case status.Code(err) == codes.Unauthenticated:
		return fmt.Errorf("entry_usecase: failed to create entry: %w: %w", entities.ErrUserTokenInvalid, err)
	case err != nil:
		return fmt.Errorf("entry_usecase: failed to get diff: %w", err)
	}
	entryMap := make(map[string]*pb.Entry)
	for _, v := range resp.Entries {
		entryMap[v.Id] = v
	}

	for _, id := range resp.DeleteIds {
		if err = uc.entryRepo.Delete(ctx, uuid.MustParse(id)); err != nil {
			uc.logger.Error("failed to delete entry", zap.Error(err))
		}
	}

	for _, v := range resp.CreateIds {
		mentry, ok := entryMap[v]
		if !ok {
			continue
		}
		now := time.Now().UTC()
		data, err := uc.encrypter.Encrypt(mentry.Data)
		if err != nil {
			return fmt.Errorf("entry_usecase: failed to encrypt entry data: %w", err)
		}
		entry := entities.Entry{
			ID:            uuid.MustParse(mentry.Id),
			Key:           mentry.Key,
			Type:          uc.toEntityType(mentry.Type),
			Meta:          mentry.Meta,
			Data:          data,
			GlobalVersion: mentry.Version,
			Version:       mentry.Version,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err = uc.entryRepo.Create(ctx, entry); err != nil {
			return fmt.Errorf("entry_usecase: failed to create entry: %w", err)
		}
	}

	for _, v := range resp.UpdateIds {
		mentry, ok := entryMap[v]
		if !ok {
			continue
		}
		data, err := uc.encrypter.Encrypt(mentry.Data)
		if err != nil {
			return fmt.Errorf("entry_usecase: failed to encrypt entry data: %w", err)
		}
		now := time.Now().UTC()
		entry := entities.Entry{
			ID:            uuid.MustParse(mentry.Id),
			Key:           mentry.Key,
			Type:          uc.toEntityType(mentry.Type),
			Meta:          mentry.Meta,
			Data:          data,
			GlobalVersion: mentry.Version,
			Version:       mentry.Version,
			UpdatedAt:     now,
		}
		if err = uc.entryRepo.Update(ctx, entry); err != nil {
			return fmt.Errorf("entry_usecase: failed to update entry: %w", err)
		}
	}

	return nil
}

func (uc *EntryUC) appendToken(ctx context.Context) (context.Context, error) {
	token, ok := uc.cache.GetString("token")
	if !ok {
		return nil, entities.ErrUserTokenNotFound
	}
	ctx = metadata.AppendToOutgoingContext(ctx, sharedmd.NewTokenKV(token)...)
	return ctx, nil
}

func (uc *EntryUC) toEntity(entry *pb.Entry) entities.Entry {
	return entities.Entry{
		ID:      uuid.MustParse(entry.Id),
		Key:     entry.Key,
		Type:    uc.toEntityType(entry.Type),
		Meta:    entry.Meta,
		Data:    entry.Data,
		Version: entry.Version,
	}
}

func (uc *EntryUC) toEntityType(t pb.EntryType) core.EntryType {
	return uc.mapper.ToEntityType(t)
}

func (uc *EntryUC) toAPIType(t core.EntryType) pb.EntryType {
	return uc.mapper.ToAPIType(t)
}
