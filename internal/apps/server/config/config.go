package config

import (
	"errors"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"time"
)

type (
	Config struct {
		Address        string        // GRPC-server address
		DatabaseDSN    string        // Database DSN
		PassHashCost   int           // Password hash cost
		TokenSecretKey []byte        // Token secret key
		TokenExpires   time.Duration // Token expires
		LogLevel       string        // Log level
		LogType        string        // Log type
		DataSecretKey  []byte        // Data secret key
		Cert           []byte
		CertKey        []byte
	}
)

func (c Config) Valid() error {
	var errs []error
	if c.Address == "" {
		errs = append(errs, errors.New("GRPC-server address should be specified"))
	}
	if c.DatabaseDSN == "" {
		errs = append(errs, errors.New("database URI should be specified"))
	}
	if c.PassHashCost < 0 {
		errs = append(errs, errors.New("password hash cost should not be negative"))
	}
	if len(c.TokenSecretKey) == 0 {
		errs = append(errs, errors.New("token secret key should be specified"))
	}
	if c.TokenExpires <= 0 {
		errs = append(errs, errors.New("token expires should be specified"))
	}
	if c.LogLevel == "" {
		errs = append(errs, errors.New("log level should be specified"))
	}
	if c.LogType == "" {
		errs = append(errs, errors.New("log type should be specified"))
	}
	if !encrypto.KeyValid(c.DataSecretKey) {
		errs = append(errs, errors.New("data secret key should be specified"))
	}
	if len(c.Cert) == 0 || len(c.CertKey) == 0 {
		errs = append(errs, errors.New("TLS certificate should be specified"))
	}
	return errors.Join(errs...)
}
