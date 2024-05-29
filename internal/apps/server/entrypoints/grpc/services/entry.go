package services

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"go.uber.org/zap"
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

func (s *EntryService) Create(_ context.Context, _ *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
	return &pb.CreateEntryResponse{}, nil
}
