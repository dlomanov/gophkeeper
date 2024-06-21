package encrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func KeyValid(key []byte) bool {
	return valid(key)
}

func Encrypt(key, data []byte) ([]byte, error) {
	if !valid(key) {
		return nil, fmt.Errorf("unsupported key length: %d, expected 16, 24 or 32", len(key))
	}
	return encrypt(key, data)
}

func Decrypt(key, encryptedData []byte) ([]byte, error) {
	if !valid(key) {
		return nil, fmt.Errorf("unsupported key length: %d, expected 16, 24 or 32", len(key))
	}
	return decrypt(key, encryptedData)
}

func encrypt(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypter: failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypter: failed to create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("encrypter: failed to generate nonce: %w", err)
	}
	cipherdata := gcm.Seal(nonce, nonce, data, nil)

	return cipherdata, nil
}

func decrypt(key, encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("encryptedData is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypter: failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypter: failed to create gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypter: data is too short")
	}
	nonce, encryptedData := encryptedData[:nonceSize], encryptedData[nonceSize:]
	data, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypter: failed to decrypt data: %w", err)
	}

	return data, nil
}

func valid(key []byte) bool {
	return len(key) == 16 || len(key) == 24 || len(key) == 32
}
