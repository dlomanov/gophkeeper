package usecases_test

import (
	"cmp"
	"context"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"slices"
	"sync"
)

var (
	_ usecases.UserRepo  = (*MockUserRepo)(nil)
	_ usecases.EntryRepo = (*MockEntryRepo)(nil)
	_ trm.Manager        = (*MockTrmManager)(nil)
)

type (
	MockUserRepo struct {
		mu      sync.RWMutex
		storage map[entities.Login]entities.User
	}
	MockEntryRepo struct {
		mu      sync.RWMutex
		storage map[string]entities.Entry
	}
	MockTrmManager struct {
	}
)

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		mu:      sync.RWMutex{},
		storage: make(map[entities.Login]entities.User),
	}
}

func (r *MockUserRepo) Exists(_ context.Context, login entities.Login) (bool, error) {
	_, ok := r.get(login)
	return ok, nil
}

func (r *MockUserRepo) Get(_ context.Context, login entities.Login) (entities.User, error) {
	user, ok := r.get(login)
	if !ok {
		return entities.User{}, entities.ErrUserNotFound
	}

	return user, nil
}

func (r *MockUserRepo) Create(_ context.Context, user entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.storage[user.Login]; ok {
		return entities.ErrUserExists
	}
	r.storage[user.Login] = user

	return nil
}

func (r *MockUserRepo) get(login entities.Login) (entities.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.storage[login]
	return user, ok
}

func NewMockEntryRepo() *MockEntryRepo {
	return &MockEntryRepo{
		mu:      sync.RWMutex{},
		storage: make(map[string]entities.Entry),
	}
}

func (r *MockEntryRepo) Get(_ context.Context, userID uuid.UUID, id uuid.UUID) (*entities.Entry, error) {
	key := r.toKey(userID, id)

	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.storage[key]
	if !ok {
		return nil, entities.ErrEntryNotFound
	}

	return &entry, nil
}

func (r *MockEntryRepo) GetAll(_ context.Context, userID uuid.UUID) ([]entities.Entry, error) {
	var entries []entities.Entry

	r.mu.RLock()
	for _, v := range r.storage {
		if v.UserID == userID {
			entries = append(entries, v)
		}
	}
	r.mu.RUnlock()

	slices.SortFunc(entries, func(a, b entities.Entry) int {
		return cmp.Compare(a.CreatedAt.Nanosecond(), b.CreatedAt.Nanosecond())
	})

	return entries, nil
}

func (r *MockEntryRepo) Create(_ context.Context, entry *entities.Entry) error {
	key := r.toKey(entry.UserID, entry.ID)

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.storage[key]; ok {
		return entities.ErrEntryExists
	}
	for _, v := range r.storage {
		if v.Key == entry.Key {
			return entities.ErrEntryExists
		}
	}
	r.storage[key] = *entry
	return nil
}

func (r *MockEntryRepo) Update(_ context.Context, entry *entities.Entry) error {
	key := r.toKey(entry.UserID, entry.ID)

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.storage[key]; !ok {
		return entities.ErrEntryNotFound
	}
	r.storage[key] = *entry
	return nil
}

func (r *MockEntryRepo) Delete(_ context.Context, userID uuid.UUID, id uuid.UUID) error {
	key := r.toKey(userID, id)

	r.mu.Lock()
	if _, ok := r.storage[key]; !ok {
		return entities.ErrEntryNotFound
	}
	delete(r.storage, key)
	r.mu.Unlock()
	return nil
}

func (r *MockEntryRepo) toKey(userID uuid.UUID, id uuid.UUID) string {
	return userID.String() + id.String()
}

func NewMockTrmManager() *MockTrmManager {
	return &MockTrmManager{}
}

func (m *MockTrmManager) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (m *MockTrmManager) DoWithSettings(
	ctx context.Context, _ trm.Settings, f func(ctx context.Context) error,
) error {
	return f(ctx)
}
