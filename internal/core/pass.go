package core

import (
	"encoding/base64"
	"fmt"
)

type (
	Pass     []byte
	PassHash []byte
	Salt     []byte
)

func NewPassHash(base64String string) (PassHash, error) {
	b, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, fmt.Errorf("pass_hash: failed to decode base64-hash: %w", err)
	}
	return b, nil
}

func NewSalt(base64String string) (Salt, error) {
	b, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, fmt.Errorf("pass_hash: failed to decode base64-salt: %w", err)
	}
	return b, nil
}

func (p Pass) Valid() bool {
	return len(p) != 0
}

func (h PassHash) Valid() bool {
	return len(h) != 0
}

func (h PassHash) Base64String() string {
	return base64.StdEncoding.EncodeToString(h)
}

func (s Salt) Base64String() string {
	return base64.StdEncoding.EncodeToString(s)
}
