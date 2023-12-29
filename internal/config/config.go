package config

import (
	"fmt"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

type Config struct {
	filepath       string
	DefaultCluster string                   `mapstructure:"default_cluster"`
	Clusters       map[string]*kafka.Config `mapstructure:"clusters"`
}

func (c *Config) Filepath() string {
	return c.filepath
}

func (c *Config) Cluster(profile string) (*kafka.Config, error) {
	if profile == "" {
		profile = c.DefaultCluster
	}

	cfg, ok := c.Clusters[profile]
	if !ok {
		return nil, fmt.Errorf("no profile specified: %s", profile)
	}

	return cfg, nil
}
