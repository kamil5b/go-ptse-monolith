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

type AsynqWorkerConfig struct {
	RedisURL       string `yaml:"redis_url"`
	Concurrency    int    `yaml:"concurrency"`
	MaxRetries     int    `yaml:"max_retries"`
	DefaultTimeout string `yaml:"default_timeout"`
}

type RabbitMQWorkerConfig struct {
	URL           string `yaml:"url"`
	Exchange      string `yaml:"exchange"`
	Queue         string `yaml:"queue"`
	WorkerCount   int    `yaml:"worker_count"`
	PrefetchCount int    `yaml:"prefetch_count"`
}

type RedpandaWorkerConfig struct {
	Brokers           []string `yaml:"brokers"`
	Topic             string   `yaml:"topic"`
	ConsumerGroup     string   `yaml:"consumer_group"`
	PartitionCount    int      `yaml:"partition_count"`
	ReplicationFactor int      `yaml:"replication_factor"`
	WorkerCount       int      `yaml:"worker_count"`
}

type WorkerConfig struct {
	Enabled  bool                 `yaml:"enabled"`
	Backend  string               `yaml:"backend"` // asynq, rabbitmq, redpanda, disable
	Asynq    AsynqWorkerConfig    `yaml:"asynq"`
	RabbitMQ RabbitMQWorkerConfig `yaml:"rabbitmq"`
	Redpanda RedpandaWorkerConfig `yaml:"redpanda"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	FromAddr string `yaml:"from_addr"`
	FromName string `yaml:"from_name"`
}

type MailgunConfig struct {
	Domain    string `yaml:"domain"`
	APIKey    string `yaml:"api_key"`
	FromAddr  string `yaml:"from_addr"`
	FromName  string `yaml:"from_name"`
	PublicKey string `yaml:"public_key"`
}

type EmailConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Provider string        `yaml:"provider"` // smtp, mailgun, noop
	SMTP     SMTPConfig    `yaml:"smtp"`
	Mailgun  MailgunConfig `yaml:"mailgun"`
}

type LocalStorageConfig struct {
	BasePath          string `yaml:"base_path"`
	MaxFileSize       int64  `yaml:"max_file_size"`
	AllowPublicAccess bool   `yaml:"allow_public_access"`
	PublicURL         string `yaml:"public_url"`
}

type S3StorageConfig struct {
	Region               string `yaml:"region"`
	Bucket               string `yaml:"bucket"`
	AccessKeyID          string `yaml:"access_key_id"`
	SecretAccessKey      string `yaml:"secret_access_key"`
	Endpoint             string `yaml:"endpoint"`
	UseSSL               bool   `yaml:"use_ssl"`
	PathStyle            bool   `yaml:"path_style"`
	PresignedURLTTL      int    `yaml:"presigned_url_ttl"`
	ServerSideEncryption bool   `yaml:"server_side_encryption"`
	StorageClass         string `yaml:"storage_class"`
}

type GCSStorageConfig struct {
	ProjectID       string `yaml:"project_id"`
	Bucket          string `yaml:"bucket"`
	CredentialsFile string `yaml:"credentials_file"`
	CredentialsJSON string `yaml:"credentials_json"`
	StorageClass    string `yaml:"storage_class"`
	Location        string `yaml:"location"`
	MetadataCache   bool   `yaml:"metadata_cache"`
}

type StorageConfig struct {
	Enabled bool               `yaml:"enabled"`
	Local   LocalStorageConfig `yaml:"local"`
	S3      S3StorageConfig    `yaml:"s3"`
	GCS     GCSStorageConfig   `yaml:"gcs"`
}

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Auth     AuthConfig     `yaml:"auth"`
	Worker   WorkerConfig   `yaml:"worker"`
	Email    EmailConfig    `yaml:"email"`
	Storage  StorageConfig  `yaml:"storage"`
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
