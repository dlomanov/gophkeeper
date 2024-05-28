package services

import (
	"context"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
)

var _ pb.EntryServiceServer = (*EntryService)(nil)

type EntryService struct {
	pb.UnimplementedEntryServiceServer
}

func NewEntryService() *EntryService {
	return &EntryService{}
}

func (s *EntryService) Create(_ context.Context, _ *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
	return &pb.CreateEntryResponse{}, nil
}
