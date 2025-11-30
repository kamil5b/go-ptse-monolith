package core

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port     string `yaml:"port"`
	GRPCPort string `yaml:"grpc_port"`
}

type SQLConfig struct {
	DBUrl string `yaml:"db_url"`
}

type DatabaseConfig struct {
	SQL   SQLConfig   `yaml:"sql"`
	Mongo MongoConfig `yaml:"mongo"`
}

type MongoConfig struct {
	MongoURL string `yaml:"mongo_url"`
	MongoDB  string `yaml:"mongo_db"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type Config struct {
	Environment string    `yaml:"environment"` // development, production
	App         AppConfig `yaml:"app"`
}

// LoadConfig loads application config from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
