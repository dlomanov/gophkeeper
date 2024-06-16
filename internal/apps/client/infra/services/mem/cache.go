package mem

import (
	"sync"
)

type (
	Cache struct {
		storage map[string]string
		rmu     sync.RWMutex
	}
)

func NewCache() *Cache {
	return &Cache{
		storage: make(map[string]string),
		rmu:     sync.RWMutex{},
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
