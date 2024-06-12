package mem

import (
	"context"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"sync"
)

type (
	Cache struct {
		storage map[string]string
		rmu     sync.RWMutex
		kvRepo  KVRepo
	}
	KVRepo interface {
		Load(ctx context.Context) ([]entities.KVPair, error)
		Upload(ctx context.Context, pairs []entities.KVPair) error
	}
)

func NewCache(
	kvRepo KVRepo,
) *Cache {
	return &Cache{
		storage: make(map[string]string),
		rmu:     sync.RWMutex{},
		kvRepo:  kvRepo,
	}
}

func (m *Cache) SetString(key string, value string) {
	m.rmu.Lock()
	defer m.rmu.Unlock()
	m.storage[key] = value
}

func (m *Cache) GetString(key string) (string, bool) {
	m.rmu.RLock()
	defer m.rmu.RUnlock()
	value, ok := m.storage[key]
	return value, ok
}

func (m *Cache) Load(ctx context.Context) error {
	m.rmu.Lock()
	defer m.rmu.Unlock()
	res, err := m.kvRepo.Load(ctx)
	if err != nil {
		return fmt.Errorf("mem: failed to load: %w", err)
	}
	for _, pair := range res {
		m.storage[pair.Key] = pair.Value
	}
	return nil
}

func (m *Cache) Flush(ctx context.Context) error {
	m.rmu.RLock()
	defer m.rmu.RUnlock()
	pairs := make([]entities.KVPair, 0, len(m.storage))
	for k, v := range m.storage {
		pairs = append(pairs, entities.KVPair{Key: k, Value: v})
	}
	if err := m.kvRepo.Upload(ctx, pairs); err != nil {
		return fmt.Errorf("mem: failed to upload: %w", err)
	}
	return nil
}
