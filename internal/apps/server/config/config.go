package config

import (
	"errors"
	"time"
)

type Config struct {
	Address        string        // GRPC-server address
	DatabaseURI    string        // Database URI
	PassHashCost   int           // Password hash cost
	TokenSecretKey string        // Token secret key
	TokenExpires   time.Duration // Token expires
	LogLevel       string        // Log level
	LogType        string        // Log type
}

func (c Config) Valid() error {
	var errs []error
	if c.Address == "" {
		errs = append(errs, errors.New("GRPC-server address should be specified"))
	}
	if c.DatabaseURI == "" {
		errs = append(errs, errors.New("database URI should be specified"))
	}
	if c.PassHashCost < 0 {
		errs = append(errs, errors.New("password hash cost should not be negative"))
	}
	if c.TokenSecretKey == "" {
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
	return errors.Join(errs...)
}
