package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

const configDirPermission = 0o755

func defaultConfig() *Config {
	return &Config{
		filepath:       "",
		DefaultCluster: "default",
		Clusters: map[string]*kafka.Config{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
			},
		},
	}
}

func LoadConfig(cfgFilepath string) (*Config, error) {
	cfgRoot := filepath.Join(xdg.ConfigHome, "kafka")

	viperConfig := viper.New()
	if cfgFilepath != "" {
		viperConfig.SetConfigFile(cfgFilepath)
	} else {
		err := os.MkdirAll(cfgRoot, configDirPermission)
		if err != nil {
			return nil, fmt.Errorf("os.MkdirAll error: %w", err)
		}

		viperConfig.SetConfigName("kafka")
		viperConfig.SetConfigType("yaml")
		viperConfig.AddConfigPath(".")
		viperConfig.AddConfigPath(cfgRoot)
	}

	err := viperConfig.ReadInConfig()
	if err != nil {
		var errNotFound viper.ConfigFileNotFoundError
		if cfgFilepath == "" && errors.As(err, &errNotFound) {
			return defaultConfig(), nil
		}

		return nil, fmt.Errorf("viper.ReadInConfig error: %w", err)
	}

	cfg := &Config{
		filepath:       viperConfig.ConfigFileUsed(),
		DefaultCluster: "",
		Clusters:       nil,
	}

	err = viperConfig.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("viper.Unmarshal error: %w", err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}
