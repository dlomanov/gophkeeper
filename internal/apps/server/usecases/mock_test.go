package usecases_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"sync"
)

var (
	_ usecases.UserRepo = (*MockUserRepo)(nil)
)

type (
	MockUserRepo struct {
		mu      sync.RWMutex
		storage map[entities.Login]entities.User
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
