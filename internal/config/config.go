package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

var (
	errClusterProfileNotFound = errors.New("cluster profile not found")
	errDefaultClusterNotSet   = errors.New("default cluster is not set")
	errNoClustersConfigured   = errors.New("no clusters configured")
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
		return nil, fmt.Errorf("%w: %s", errClusterProfileNotFound, profile)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.DefaultCluster) == "" {
		return errDefaultClusterNotSet
	}

	if len(c.Clusters) == 0 {
		return errNoClustersConfigured
	}

	for name, cluster := range c.Clusters {
		if cluster == nil {
			return fmt.Errorf("cluster %q: %w", name, kafka.ErrConfigRequired)
		}

		err := cluster.Validate()
		if err != nil {
			return fmt.Errorf("cluster %q: %w", name, err)
		}
	}

	if _, ok := c.Clusters[c.DefaultCluster]; !ok {
		return fmt.Errorf("%w: %s", errClusterProfileNotFound, c.DefaultCluster)
	}

	return nil
}
