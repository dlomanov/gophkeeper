package mem

import (
	"context"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
)

type (
	Storage struct {
		kvRepo KVRepo
	}
	KVRepo interface {
		Load(ctx context.Context) ([]entities.KVPair, error)
		Upload(ctx context.Context, pairs []entities.KVPair) error
	}
)

func NewStorage(kvRepo KVRepo) *Storage {
	return &Storage{
		kvRepo: kvRepo,
	}
}

func (s *Storage) Load(ctx context.Context, c *Cache) error {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	if s.kvRepo == nil {
		return fmt.Errorf("mem: KV-repo is nil")
	}

	res, err := s.kvRepo.Load(ctx)
	if err != nil {
		return fmt.Errorf("mem: failed to load: %w", err)
	}
	for _, pair := range res {
		c.storage[pair.Key] = pair.Value
	}
	return nil
}

func (s *Storage) Flush(ctx context.Context, c *Cache) error {
	c.rmu.RLock()
	defer c.rmu.RUnlock()
	if s.kvRepo == nil {
		return fmt.Errorf("mem: KV-repo is nil")
	}

	pairs := make([]entities.KVPair, 0, len(c.storage))
	for k, v := range c.storage {
		pairs = append(pairs, entities.KVPair{Key: k, Value: v})
	}
	if err := s.kvRepo.Upload(ctx, pairs); err != nil {
		return fmt.Errorf("mem: failed to upload: %w", err)
	}
	return nil
}
