package core

import (
	"os"

	"gopkg.in/yaml.v3"
)

type HandlerFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, v1
	Product        string `yaml:"product"`        // disable, v1
	User           string `yaml:"user"`           // disable, v1
}

type ServiceFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, v1
	Product        string `yaml:"product"`        // disable, v1
	User           string `yaml:"user"`           // disable, v1
}

type RepositoryFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, postgres, mongo
	Product        string `yaml:"product"`        // disable, postgres, mongo
	User           string `yaml:"user"`           // disable, postgres, mongo
}

type WorkerTaskFeatureFlag struct {
	EmailNotifications bool `yaml:"email_notifications"`
	DataExport         bool `yaml:"data_export"`
	ReportGeneration   bool `yaml:"report_generation"`
	ImageProcessing    bool `yaml:"image_processing"`
}

type WorkerFeatureFlag struct {
	Enabled bool                  `yaml:"enabled"`
	Backend string                `yaml:"backend"` // asynq, rabbitmq, redpanda, disable
	Tasks   WorkerTaskFeatureFlag `yaml:"tasks"`
}

type EmailFeatureFlag struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"` // smtp, mailgun, noop
}

type StorageS3FeatureFlag struct {
	EnableEncryption bool   `yaml:"enable_encryption"`
	StorageClass     string `yaml:"storage_class"`
	PresignedURLTTL  int    `yaml:"presigned_url_ttl"`
}

type StorageGCSFeatureFlag struct {
	StorageClass  string `yaml:"storage_class"`
	MetadataCache bool   `yaml:"metadata_cache"`
}

type StorageFeatureFlag struct {
	Enabled bool                  `yaml:"enabled"`
	Backend string                `yaml:"backend"` // local, s3, gcs, s3-compatible, noop
	S3      StorageS3FeatureFlag  `yaml:"s3"`
	GCS     StorageGCSFeatureFlag `yaml:"gcs"`
}

type FeatureFlag struct {
	HTTPHandler string `yaml:"http_handler"` // echo, gin
	Cache       string `yaml:"cache"`        // redis, memory, disable

	Handler    HandlerFeatureFlag    `yaml:"handler"`
	Service    ServiceFeatureFlag    `yaml:"service"`
	Repository RepositoryFeatureFlag `yaml:"repository"`
	Worker     WorkerFeatureFlag     `yaml:"worker"`
	Email      EmailFeatureFlag      `yaml:"email"`
	Storage    StorageFeatureFlag    `yaml:"storage"`
}

// LoadFeatureFlags loads feature flag configuration from a YAML file.
func LoadFeatureFlags(path string) (*FeatureFlag, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ff FeatureFlag
	if err := yaml.Unmarshal(data, &ff); err != nil {
		return nil, err
	}
	return &ff, nil
}
