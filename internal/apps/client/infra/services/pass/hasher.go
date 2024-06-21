package pass

import (
	"crypto/rand"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/core"
	"golang.org/x/crypto/scrypt"
)

type Hasher struct {
}

func (h Hasher) Hash(pass core.Pass, salt core.Salt) (core.PassHash, error) {
	return h.hash(pass, salt)
}

func (h Hasher) Compare(pass core.Pass, salt core.Salt, hash core.PassHash) bool {
	hash2, err := h.hash(pass, salt)
	if err != nil {
		return false
	}
	if len(hash) != len(hash2) {
		return false
	}
	for i := range hash {
		if hash[i] != hash2[i] {
			return false
		}
	}
	return true
}

func (h Hasher) GenerateSalt() (core.Salt, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("hasher: failed to generate salt: %w", err)
	}
	return salt, nil
}

func (h Hasher) hash(pass core.Pass, salt core.Salt) (core.PassHash, error) {
	key, err := scrypt.Key(pass, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("hasher: failed to hash pass: %w", err)
	}
	return key, nil
}
