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

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	MaxRetries   int    `yaml:"max_retries"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type JWTConfig struct {
	Secret               string `yaml:"secret"`
	AccessTokenDuration  string `yaml:"access_token_duration"`
	RefreshTokenDuration string `yaml:"refresh_token_duration"`
}

type AuthConfig struct {
	Type          string `yaml:"type"`           // jwt, session, basic, none
	SessionCookie string `yaml:"session_cookie"` // cookie name for session-based auth
	BcryptCost    int    `yaml:"bcrypt_cost"`    // bcrypt cost for password hashing
}

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Auth     AuthConfig     `yaml:"auth"`
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
