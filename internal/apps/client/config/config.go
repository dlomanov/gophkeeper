package config

import "errors"

type Config struct {
	Address        string   // GRPC-server address
	LogLevel       string   // log level
	LogType        string   // logger type
	LogOutputPaths []string // logger output paths
	Cert           []byte   // TLS certificate
	DSN            string   // database DSN
	BuildVersion   string   // build version info
	BuildDate      string   // build date info
	BuildCommit    string   // build commit info
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
	if c.DSN == "" {
		errs = append(errs, errors.New("database DSN should be specified"))
	}
	return errors.Join(errs...)
}
