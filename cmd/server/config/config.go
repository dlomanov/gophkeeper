package config

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	"gopkg.in/yaml.v2"
)

type config struct {
	Address    string `yaml:"address" env:"ADDRESS"`
	ConfigPath string `yaml:"config,omitempty" env:"CONFIG"`
}

//go:embed config.yaml
var configFS embed.FS

func Parse() server.Config {
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
	flag.StringVar(&c.Address, "a", c.Address, "GRPC-server address (shorthand)")
	flag.StringVar(&c.Address, "address", c.Address, "GRPC-server address")
	flag.StringVar(&c.ConfigPath, "c", c.ConfigPath, "config path (shorthand)")
	flag.StringVar(&c.ConfigPath, "config", c.ConfigPath, "config path")
	flag.Parse()
}

func (c *config) readEnv() {
	err := env.Parse(c)
	if err != nil {
		panic(err)
	}
}

func (c config) print() {
	content, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))
}

func (c *config) toConfig() server.Config {
	return server.Config{
		Address: c.Address,
	}
}
