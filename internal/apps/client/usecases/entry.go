package usecases

import (
	"context"
	"fmt"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	EntryUC struct {
		logger      *zap.Logger
		cache       *cache.Cache
		entryClient pb.EntryServiceClient
	}
	GetEntriesResponse struct {
		Entries []entities.Entry
	}
	CreateEntryRequest struct {
		Key  string
		Type entities.EntryType
		Meta map[string]string
		Data []byte
	}
	CreateEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
	UpdateEntryRequest struct {
		ID      uuid.UUID
		Meta    map[string]string
		Data    []byte
		Version int64
	}
	UpdateEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
	DeleteEntryRequest struct {
		ID uuid.UUID
	}
	DeleteEntryResponse struct {
		ID      uuid.UUID
		Version int64
	}
)

func NewEntriesUC(
	logger *zap.Logger,
	cache *cache.Cache,
	entryClient pb.EntryServiceClient,
) *EntryUC {
	return &EntryUC{
		logger:      logger,
		cache:       cache,
		entryClient: entryClient,
	}
}

func (uc *EntryUC) GetAll(ctx context.Context) (response GetEntriesResponse, err error) {
	if ctx, err = uc.appendToken(ctx); err != nil {
		return response, err
	}
	resp, err := uc.entryClient.GetAll(ctx, &pb.GetEntriesRequest{})
	if err != nil {
		return response, fmt.Errorf("entry_usecase: failed to get entries: %w", err)
	}
	entries := make([]entities.Entry, len(resp.Entries))
	for i, entry := range resp.Entries {
		entries[i] = uc.toEntity(entry)
	}
	response.Entries = entries
	return response, err
}

func (uc *EntryUC) Create(
	ctx context.Context,
	request CreateEntryRequest,
) (response CreateEntryResponse, err error) {
	if ctx, err = uc.appendToken(ctx); err != nil {
		return response, err
	}
	resp, err := uc.entryClient.Create(ctx, &pb.CreateEntryRequest{
		Key:  request.Key,
		Type: uc.toAPIType(request.Type),
		Meta: request.Meta,
		Data: request.Data,
	})
	switch {
	case status.Code(err) == codes.AlreadyExists:
		return response, entities.ErrEntryExists
	case status.Code(err) == codes.InvalidArgument:
		return response, entities.ErrEntryInvalid
	case err != nil:
		return response, fmt.Errorf("entry_usecase: failed to create entry: %w", err)
	}
	response.ID = uuid.MustParse(resp.Id)
	response.Version = resp.Version
	return response, nil
}

func (uc *EntryUC) Update(
	ctx context.Context,
	request UpdateEntryRequest,
) (response UpdateEntryResponse, err error) {
	if ctx, err = uc.appendToken(ctx); err != nil {
		return response, err
	}
	resp, err := uc.entryClient.Update(ctx, &pb.UpdateEntryRequest{
		Id:      request.ID.String(),
		Meta:    request.Meta,
		Data:    request.Data,
		Version: request.Version,
	})
	switch {
	case status.Code(err) == codes.NotFound:
		return response, entities.ErrEntryNotFound
	case status.Code(err) == codes.InvalidArgument:
		return response, entities.ErrEntryInvalid
	case err != nil:
		return response, fmt.Errorf("entry_usecase: failed to update entry: %w", err)
	}
	response.ID = uuid.MustParse(resp.Id)
	response.Version = resp.Version
	return response, nil
}

func (uc *EntryUC) Delete(
	ctx context.Context,
	request DeleteEntryRequest,
) (response DeleteEntryResponse, err error) {
	if ctx, err = uc.appendToken(ctx); err != nil {
		return response, err
	}
	resp, err := uc.entryClient.Delete(ctx, &pb.DeleteEntryRequest{Id: request.ID.String()})
	switch {
	case status.Code(err) == codes.NotFound:
		return response, entities.ErrEntryNotFound
	case status.Code(err) == codes.InvalidArgument:
		return response, entities.ErrEntryInvalid
	case err != nil:
		return response, fmt.Errorf("entry_usecase: failed to delete entry: %w", err)
	}
	response.ID = uuid.MustParse(resp.Id)
	response.Version = resp.Version
	return response, nil
}

func (uc *EntryUC) appendToken(ctx context.Context) (context.Context, error) {
	v, ok := uc.cache.Get("token")
	if !ok {
		return nil, entities.ErrUserTokenNotFound
	}
	token := v.(string)
	ctx = metadata.AppendToOutgoingContext(ctx, sharedmd.NewTokenKV(token)...)
	return ctx, nil
}

func (uc *EntryUC) toEntity(entry *pb.Entry) entities.Entry {
	return entities.Entry{
		ID: uuid.MustParse(entry.Id),
		//UserID:  uuid.Nil,
		Key:     entry.Key,
		Type:    uc.toEntityType(entry.Type),
		Meta:    entry.Meta,
		Data:    entry.Data,
		Version: entry.Version,
		//CreatedAt
		//UpdatedAt
	}
}

func (uc *EntryUC) toEntityType(t pb.EntryType) entities.EntryType {
	switch t {
	case pb.EntryType_ENTRY_TYPE_UNSPECIFIED:
		return entities.EntryTypeUnspecified
	case pb.EntryType_ENTRY_TYPE_PASSWORD:
		return entities.EntryTypePassword
	case pb.EntryType_ENTRY_TYPE_NOTE:
		return entities.EntryTypeNote
	case pb.EntryType_ENTRY_TYPE_CARD:
		return entities.EntryTypeCard
	case pb.EntryType_ENTRY_TYPE_BINARY:
		return entities.EntryTypeBinary
	default:
		return entities.EntryTypeUnspecified
	}
}

func (uc *EntryUC) toAPIType(t entities.EntryType) pb.EntryType {
	switch t {
	case entities.EntryTypeUnspecified:
		return pb.EntryType_ENTRY_TYPE_UNSPECIFIED
	case entities.EntryTypePassword:
		return pb.EntryType_ENTRY_TYPE_PASSWORD
	case entities.EntryTypeNote:
		return pb.EntryType_ENTRY_TYPE_NOTE
	case entities.EntryTypeCard:
		return pb.EntryType_ENTRY_TYPE_CARD
	case entities.EntryTypeBinary:
		return pb.EntryType_ENTRY_TYPE_BINARY
	default:
		return pb.EntryType_ENTRY_TYPE_UNSPECIFIED
	}
}
