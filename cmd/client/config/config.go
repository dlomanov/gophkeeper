package config

import (
	"embed"
	"flag"
	"fmt"
	clientcfg "github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env"
	"gopkg.in/yaml.v3"
)

type config struct {
	Address        string `yaml:"address" env:"ADDRESS"`
	ConfigPath     string `yaml:"config,omitempty" env:"CONFIG"`
	LogLevel       string `yaml:"log_level" env:"LOG_LEVEL"`
	LogType        string `yaml:"log_type" env:"LOG_TYPE"`
	LogOutputPaths string `yaml:"log_output_paths" env:"LOG_OUTPUT_PATHS"`
	CertPath       string `yaml:"cert_path" env:"CERT_PATH"`
}

//go:embed config.yaml
var configFS embed.FS

func Parse(print bool) clientcfg.Config {
	c := &config{}
	c.readDefaults()
	c.readConfig()
	c.readFlags()
	c.readEnv()
	if print {
		c.print()
	}
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
	flag.StringVar(&c.Address, "address", c.Address, "GRPC-server address")
	flag.StringVar(&c.ConfigPath, "config", c.ConfigPath, "config path")
	flag.StringVar(&c.LogLevel, "log_level", c.LogLevel, "log level")
	flag.StringVar(&c.LogType, "log_type", c.LogType, "log type")
	flag.StringVar(&c.LogOutputPaths, "log_output_paths", c.LogOutputPaths, "log output paths")
	flag.StringVar(&c.CertPath, "cert_path", c.CertPath, "cert path")

	flag.Parse()
}

func (c *config) readEnv() {
	err := env.Parse(c)
	if err != nil {
		panic(err)
	}
}

func (c config) print() {
	c.CertPath = "**********"
	content, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))
}

func (c *config) toConfig() clientcfg.Config {
	cert := c.readCert()
	return clientcfg.Config{
		Address:        c.Address,
		LogLevel:       c.LogLevel,
		LogType:        c.LogType,
		LogOutputPaths: c.parseLogOutputPaths(),
		Cert:           cert,
	}
}

func (c *config) parseLogOutputPaths() []string {
	if c.LogOutputPaths == "" {
		return nil
	}
	paths := strings.Split(c.LogOutputPaths, ",")
	if len(paths) == 0 {
		return nil
	}
	return paths
}

func (c *config) readCert() (cert []byte) {
	if c.CertPath == "" {
		return nil
	}

	cert, err := os.ReadFile(c.CertPath)
	if err != nil {
		log.Fatalf("failed to read TLS-certificate: %v", err)
	}
	return cert
}
