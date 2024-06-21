package pass

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/core"
	"golang.org/x/crypto/bcrypt"
)

var (
	_ usecases.PassHasher = (*Hasher)(nil)

	empty core.PassHash
)

type Hasher struct {
	cost int
}

func NewHasher(cost int) Hasher {
	return Hasher{cost: cost}
}

func (h Hasher) Hash(password core.Pass) (core.PassHash, error) {
	hash, err := bcrypt.GenerateFromPassword(password, h.cost)
	if err != nil {
		return empty, err
	}
	return hash, nil
}

func (h Hasher) Compare(pass core.Pass, hash core.PassHash) bool {
	err := bcrypt.CompareHashAndPassword(hash, pass)
	return err == nil
}
