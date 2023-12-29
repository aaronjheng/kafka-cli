package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

var defaultConfig = &Config{
	DefaultCluster: "default",
	Clusters: map[string]*kafka.Config{
		"default": {
			Brokers: []string{"127.0.0.1:9092"},
		},
	},
}

func LoadConfig(cfgFilepath string) (*Config, error) {
	cfgRoot := path.Join(xdg.ConfigHome, "kafka")
	if _, err := os.Stat(cfgRoot); os.IsNotExist(err) {
		if err := os.Mkdir(cfgRoot, 0o755); err != nil {
			return nil, fmt.Errorf("os.Mkdir error: %w", err)
		}
	}

	if cfgFilepath != "" {
		viper.SetConfigFile(cfgFilepath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(cfgRoot)
	}

	if err := viper.ReadInConfig(); err != nil {
		var errNotFound viper.ConfigFileNotFoundError
		if cfgFilepath == "" && errors.As(err, &errNotFound) {
			return defaultConfig, nil
		} else {
			return nil, fmt.Errorf("viper.ReadInConfig error: %w", err)
		}
	}

	cfg := &Config{
		filepath: viper.GetViper().ConfigFileUsed(),
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal error: %w", err)
	}

	return cfg, nil
}
