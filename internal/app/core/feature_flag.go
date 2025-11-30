package core

import (
	"os"

	"gopkg.in/yaml.v3"
)

type HandlerFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, v1
	Product        string `yaml:"product"`        // disable, v1
}

type ServiceFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, v1
	Product        string `yaml:"product"`        // disable, v1
}

type RepositoryFeatureFlag struct {
	Authentication string `yaml:"authentication"` // disable, postgres, mongo
	Product        string `yaml:"product"`        // disable, postgres, mongo
}

type FeatureFlag struct {
	HTTPHandler string `yaml:"http_handler"` // echo, gin

	Handler    HandlerFeatureFlag    `yaml:"handler"`
	Service    ServiceFeatureFlag    `yaml:"service"`
	Repository RepositoryFeatureFlag `yaml:"repository"`
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
