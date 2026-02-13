package config

import (
	"errors"
	"fmt"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

var errClusterProfileNotFound = errors.New("cluster profile not found")

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
		return nil, fmt.Errorf("%w: %s", errClusterProfileNotFound, profile)
	}

	return cfg, nil
}
