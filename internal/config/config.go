package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

func New() (*Config, error) {
	cfg := &Config{
		AppEnv: os.Getenv("APP_ENV"),
	}
	yamlFile, err := os.ReadFile(fmt.Sprintf("./configs/%s.yml", cfg.AppEnviroment()))
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type Config struct {
	AppEnv string `env:"APP_ENV"`
	Badger struct {
		Path string `yaml:"path"`
	} `yaml:"badger"`
	Server struct {
		Port                          int           `yaml:"port"`
		ReadTimeout                   time.Duration `yaml:"read_timeout"`
		WriteTimeout                  time.Duration `yaml:"write_timeout"`
		IdleTimeout                   time.Duration `yaml:"idle_timeout"`
		ShutdownTimeout               time.Duration `yaml:"shutdown_timeout"`
		LimitConcurrentFileRequest    int           `yaml:"limit_concurrent_file_request"`
		LimitConcurrentListingRequest int           `yaml:"limit_concurrent_listing_request"`
	} `yaml:"server"`
}

type Jaeger struct {
	Endpoint    string  `yaml:"endpoint"`
	ServiceName string  `yaml:"service_name"`
	SampleRatio float64 `yaml:"sample_ratio"`
}

func (c *Config) AppEnviroment() string {
	if c.AppEnv == "" {
		return "local"
	}
	return c.AppEnv
}
