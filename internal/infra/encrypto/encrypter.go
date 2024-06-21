package encrypto

import "fmt"

type Encrypter struct {
	Key []byte
}

func NewEncrypter(key []byte) (*Encrypter, error) {
	if !valid(key) {
		return nil, fmt.Errorf("crypto: unsupported key length: %d, expected 16, 24 or 32", len(key))
	}
	return &Encrypter{Key: key}, nil
}

func (e *Encrypter) Encrypt(data []byte) ([]byte, error) {
	return encrypt(e.Key, data)
}

func (e *Encrypter) Decrypt(encryptedData []byte) ([]byte, error) {
	return decrypt(e.Key, encryptedData)
}
