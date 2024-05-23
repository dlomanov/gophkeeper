package pass

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

var (
	_ usecases.PassHasher = (*Hasher)(nil)

	empty entities.PassHash = ""
)

type Hasher struct {
	cost int
}

func NewHasher(cost int) Hasher {
	return Hasher{cost: cost}
}

func (h Hasher) Hash(password entities.Pass) (entities.PassHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return empty, err
	}
	return entities.PassHash(hash), nil
}

func (h Hasher) Compare(password entities.Pass, hash entities.PassHash) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
