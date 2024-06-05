package services

import (
	"context"
	"errors"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc/interceptor"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/core/apperrors"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.EntryServiceServer = (*EntryService)(nil)

type EntryService struct {
	pb.UnimplementedEntryServiceServer
	logger  *zap.Logger
	entryUC *usecases.EntryUC
}

func NewEntryService(
	logger *zap.Logger,
	entryUC *usecases.EntryUC,
) *EntryService {
	return &EntryService{
		logger:  logger,
		entryUC: entryUC,
	}
}

func (s *EntryService) Get(
	ctx context.Context,
	request *pb.GetEntryRequest,
) (*pb.GetEntryResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}
	id, err := uuid.Parse(request.Id)
	if err != nil {
		s.logger.Debug("invalid entry id", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, entities.ErrEntryIDInvalid.Error())
	}

	got, err := s.entryUC.Get(ctx, usecases.GetEntryRequest{UserID: userID, ID: id})
	var (
		invalid  *apperrors.AppErrorInvalid
		notFound *apperrors.AppErrorNotFound
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &notFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		s.logger.Error("failed to get entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", id.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.GetEntryResponse{Entry: s.toAPIEntry(*got.Entry)}, nil
}

func (s *EntryService) GetAll(
	ctx context.Context,
	_ *pb.GetAllEntriesRequest,
) (*pb.GetAllEntriesResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}

	got, err := s.entryUC.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID})
	var (
		invalid *apperrors.AppErrorInvalid
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		s.logger.Error("failed to get entries",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	entries := make([]*pb.Entry, len(got.Entries))
	for i, entry := range got.Entries {
		entries[i] = s.toAPIEntry(entry)
	}
	return &pb.GetAllEntriesResponse{Entries: entries}, nil
}

func (s *EntryService) Create(
	ctx context.Context,
	request *pb.CreateEntryRequest,
) (*pb.CreateEntryResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}

	created, err := s.entryUC.Create(ctx, usecases.CreateEntryRequest{
		Key:    request.Key,
		UserID: userID,
		Type:   s.toEntityType(request.Type),
		Meta:   request.Meta,
		Data:   request.Data,
	})
	var (
		invalid  *apperrors.AppErrorInvalid
		conflict *apperrors.AppErrorConflict
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &conflict):
		return nil, status.Error(codes.AlreadyExists, err.Error())
	case err != nil:
		s.logger.Error("failed to create entry",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.CreateEntryResponse{
		Id:      created.ID.String(),
		Version: created.Version,
	}, nil
}

func (s *EntryService) Update(
	ctx context.Context,
	request *pb.UpdateEntryRequest,
) (*pb.UpdateEntryResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}

	updated, err := s.entryUC.Update(ctx, usecases.UpdateEntryRequest{
		ID:      s.parseUUID(request.Id),
		UserID:  userID,
		Meta:    request.Meta,
		Data:    request.Data,
		Version: request.Version,
	})
	var (
		invalid  *apperrors.AppErrorInvalid
		notFound *apperrors.AppErrorNotFound
		conflict *apperrors.AppErrorConflict
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &notFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.As(err, &conflict):
		return nil, status.Error(codes.AlreadyExists, err.Error())
	case err != nil:
		s.logger.Error("failed to update entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", request.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.UpdateEntryResponse{
		Id:      updated.ID.String(),
		Version: updated.Version,
	}, nil
}

func (s *EntryService) UpdateForced(
	ctx context.Context,
	request *pb.UpdateEntryRequest,
) (*pb.UpdateEntryResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}

	updated, err := s.entryUC.UpdateForced(ctx, usecases.UpdateEntryRequest{
		ID:      s.parseUUID(request.Id),
		UserID:  userID,
		Meta:    request.Meta,
		Data:    request.Data,
		Version: request.Version,
	})
	var (
		invalid  *apperrors.AppErrorInvalid
		notFound *apperrors.AppErrorNotFound
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &notFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		s.logger.Error("failed to update entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", request.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.UpdateEntryResponse{
		Id:      updated.ID.String(),
		Version: updated.Version,
	}, nil
}

func (s *EntryService) Delete(
	ctx context.Context,
	request *pb.DeleteEntryRequest,
) (*pb.DeleteEntryResponse, error) {
	userID, ok := interceptor.GetUserID(ctx)
	if !ok {
		s.logger.Debug("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, entities.ErrUserIDInvalid.Error())
	}

	deleted, err := s.entryUC.Delete(ctx, usecases.DeleteEntryRequest{
		ID:     s.parseUUID(request.Id),
		UserID: userID,
	})
	var (
		invalid  *apperrors.AppErrorInvalid
		notFound *apperrors.AppErrorNotFound
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &notFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		s.logger.Error("failed to update entry",
			zap.String("user_id", userID.String()),
			zap.String("entry_id", request.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.DeleteEntryResponse{
		Id:      deleted.ID.String(),
		Version: deleted.Version,
	}, nil
}

func (s *EntryService) toAPIEntry(entry entities.Entry) *pb.Entry {
	return &pb.Entry{
		Id:      entry.ID.String(),
		Key:     entry.Key,
		Type:    s.toAPIType(entry.Type),
		Meta:    entry.Meta,
		Data:    entry.Data,
		Version: entry.Version,
	}
}

func (s *EntryService) toAPIType(typ entities.EntryType) pb.EntryType {
	switch typ {
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

func (s *EntryService) toEntityType(typ pb.EntryType) entities.EntryType {
	switch typ {
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

func (s *EntryService) parseUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		s.logger.Debug("failed to parse uuid",
			zap.String("value", value),
			zap.Error(err))
		return uuid.Nil
	}
	return id
}
