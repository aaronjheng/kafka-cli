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

const configDirPermission = 0o755

//nolint:gochecknoglobals // Shared immutable fallback config.
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

	_, err := os.Stat(cfgRoot)
	if os.IsNotExist(err) {
		err = os.Mkdir(cfgRoot, configDirPermission)
		if err != nil {
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

	err = viper.ReadInConfig()
	if err != nil {
		var errNotFound viper.ConfigFileNotFoundError
		if cfgFilepath == "" && errors.As(err, &errNotFound) {
			return defaultConfig, nil
		}

		return nil, fmt.Errorf("viper.ReadInConfig error: %w", err)
	}

	cfg := &Config{
		filepath: viper.GetViper().ConfigFileUsed(),
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("viper.Unmarshal error: %w", err)
	}

	return cfg, nil
}
