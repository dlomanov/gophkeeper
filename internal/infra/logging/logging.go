package logging

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	LogTypeDevelopment LogType = "development"
	LogTypeProduction  LogType = "production"
)

type (
	Config struct {
		Level string
		Type  string
	}
	LogType string
)

func NewLogger(config Config) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return nil, err
	}

	var c zap.Config
	switch LogType(config.Type) {
	case LogTypeDevelopment:
		c = zap.NewDevelopmentConfig()
	case LogTypeProduction:
		c = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("unknown logger type %s", config.Type)
	}

	c.Level = lvl
	return c.Build()
}
