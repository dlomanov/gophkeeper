package config

import "errors"

type Config struct {
	Address        string // GRPC-server address
	LogLevel       string
	LogType        string
	LogOutputPaths []string
	Cert           []byte
}

func (c Config) Validate() error {
	var errs []error
	if c.Address == "" {
		errs = append(errs, errors.New("GRPC-server address should be specified"))
	}
	if c.LogLevel == "" {
		errs = append(errs, errors.New("log level should be specified"))
	}
	if c.LogType == "" {
		errs = append(errs, errors.New("log type should be specified"))
	}
	if len(c.Cert) == 0 {
		errs = append(errs, errors.New("TLS certificate should be specified"))
	}
	return errors.Join(errs...)
}
