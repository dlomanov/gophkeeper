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
	DatabaseURI    string        `yaml:"database_uri" env:"DATABASE_URI"`
	PassHashCost   int           `yaml:"pass_hash_cost" env:"PASS_HASH_COST"`
	TokenSecretKey string        `yaml:"token_secret_key" env:"TOKEN_SECRET_KEY"`
	TokenExpires   time.Duration `yaml:"token_expires" env:"TOKEN_EXPIRES"`
	LogLevel       string        `yaml:"log_level" env:"LOG_LEVEL"`
	LogType        string        `yaml:"log_type" env:"LOG_TYPE"`
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
		panic(err)
	}
	if err := yaml.Unmarshal(content, c); err != nil {
		panic(err)
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
		panic(err)
	}
	err = yaml.Unmarshal(content, c)
	if err != nil {
		panic(err)
	}
}

func (c *config) readFlags() {
	flag.StringVar(&c.ConfigPath, "c", c.ConfigPath, "config path (shorthand)")
	flag.StringVar(&c.ConfigPath, "config", c.ConfigPath, "config path")
	flag.StringVar(&c.Address, "address", c.Address, "GRPC-server address")
	flag.StringVar(&c.DatabaseURI, "database_uri", c.DatabaseURI, "database URI")
	flag.IntVar(&c.PassHashCost, "pass_hash_cost", c.PassHashCost, "password hash cost")
	flag.StringVar(&c.TokenSecretKey, "token_secret_key", c.TokenSecretKey, "token secret key")
	flag.DurationVar(&c.TokenExpires, "token_expires", c.TokenExpires, "token expires")
	flag.StringVar(&c.LogLevel, "log_level", c.LogLevel, "log level")
	flag.StringVar(&c.LogType, "log_type", c.LogType, "log type")
	flag.Parse()
}

func (c *config) readEnv() {
	err := env.Parse(c)
	if err != nil {
		panic(err)
	}
}

func (c config) print() {
	c.TokenSecretKey = "**********"
	content, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))
}

func (c *config) toConfig() *srvcfg.Config {
	res := &srvcfg.Config{
		Address:        c.Address,
		DatabaseURI:    c.DatabaseURI,
		PassHashCost:   c.PassHashCost,
		TokenSecretKey: c.TokenSecretKey,
		TokenExpires:   c.TokenExpires,
		LogLevel:       c.LogLevel,
		LogType:        c.LogType,
	}
	if err := res.Valid(); err != nil {
		log.Fatal(err)
	}
	return res
}
