package config

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env"
	srvcfg "github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"gopkg.in/yaml.v2"
)

type config struct {
	Address        string        `yaml:"address" env:"ADDRESS"`
	ConfigPath     string        `yaml:"config,omitempty" env:"CONFIG"`
	DatabaseDSN    string        `yaml:"database_dsn" env:"DATABASE_DSN"`
	PassHashCost   int           `yaml:"pass_hash_cost" env:"PASS_HASH_COST"`
	TokenSecretKey string        `yaml:"token_secret_key" env:"TOKEN_SECRET_KEY"`
	TokenExpires   time.Duration `yaml:"token_expires" env:"TOKEN_EXPIRES"`
	LogLevel       string        `yaml:"log_level" env:"LOG_LEVEL"`
	LogType        string        `yaml:"log_type" env:"LOG_TYPE"`
	DataSecretKey  string        `yaml:"data_secret_key" env:"DATA_SECRET_KEY"`
	CertPath       string        `yaml:"cert_path" env:"CERT_PATH"`
	CertKeyPath    string        `yaml:"cert_key_path" env:"CERT_KEY_PATH"`
}

//go:embed config.yaml
var configFS embed.FS

func Parse() *srvcfg.Config {
	c := &config{}
	c.readDefaults()
	c.readConfig()
	c.readFlags()
	c.readEnv()
	c.print()
	return c.toConfig()
}

func (c *config) readDefaults() {
	content, err := configFS.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("failed to read default config: %v", err)
	}
	if err := yaml.Unmarshal(content, c); err != nil {
		log.Fatalf("failed to unmarshal default config: %v", err)
	}
}

func (c *config) readConfig() {
	path := ""

	v := flag.Lookup("c")
	if v == nil {
		v = flag.Lookup("config")
	}
	if v != nil {
		path = v.Value.String()
	}
	cp, ok := os.LookupEnv("CONFIG")
	if ok {
		path = cp
	}
	if path == "" {
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	err = yaml.Unmarshal(content, c)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}
}

func (c *config) readFlags() {
	flag.StringVar(&c.ConfigPath, "c", c.ConfigPath, "config path (shorthand)")
	flag.StringVar(&c.ConfigPath, "config", c.ConfigPath, "config path")
	flag.StringVar(&c.Address, "address", c.Address, "GRPC-server address")
	flag.StringVar(&c.DatabaseDSN, "database_dsn", c.DatabaseDSN, "database DSN")
	flag.IntVar(&c.PassHashCost, "pass_hash_cost", c.PassHashCost, "password hash cost")
	flag.StringVar(&c.TokenSecretKey, "token_secret_key", c.TokenSecretKey, "token secret key")
	flag.DurationVar(&c.TokenExpires, "token_expires", c.TokenExpires, "token expires")
	flag.StringVar(&c.LogLevel, "log_level", c.LogLevel, "log level")
	flag.StringVar(&c.LogType, "log_type", c.LogType, "log type")
	flag.StringVar(&c.DataSecretKey, "data_secret_key", c.DataSecretKey, "data secret key 16/24/32 bytes")
	flag.StringVar(&c.CertPath, "cert_path", c.CertPath, "TLS-certificate file path")
	flag.StringVar(&c.CertKeyPath, "cert_key_path", c.CertKeyPath, "TLS-certificate key file path")
	flag.Parse()
}

func (c *config) readEnv() {
	err := env.Parse(c)
	if err != nil {
		log.Fatalf("failed to read env: %v", err)
	}
}

func (c config) print() {
	c.TokenSecretKey = "**********"
	c.DataSecretKey = "**********"
	c.CertPath = "**********"
	c.CertKeyPath = "**********"
	content, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("failed to print config: %v", err)
	}
	fmt.Println(string(content))
}

func (c *config) toConfig() *srvcfg.Config {
	cert, certKey := c.readCert()

	res := &srvcfg.Config{
		Address:        c.Address,
		DatabaseDSN:    c.DatabaseDSN,
		PassHashCost:   c.PassHashCost,
		TokenSecretKey: []byte(c.TokenSecretKey),
		TokenExpires:   c.TokenExpires,
		LogLevel:       c.LogLevel,
		LogType:        c.LogType,
		DataSecretKey:  []byte(c.DataSecretKey),
		Cert:           cert,
		CertKey:        certKey,
	}
	return res
}

func (c *config) readCert() (cert, certKey []byte) {
	if c.CertPath == "" || c.CertKeyPath == "" {
		return nil, nil
	}

	cert, err := os.ReadFile(c.CertPath)
	if err != nil {
		log.Fatalf("failed to read TLS-certificate: %v", err)
	}
	certKey, err = os.ReadFile(c.CertKeyPath)
	if err != nil {
		log.Fatalf("failed to read TLS-certificate key: %v", err)
	}
	return cert, certKey
}
