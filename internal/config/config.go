package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"path"

	"github.com/Shopify/sarama"
	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type Config struct {
	filepath       string
	DefaultCluster string              `mapstructure:"default_cluster"`
	Clusters       map[string]*Cluster `mapstructure:"clusters"`
}

func (c *Config) Filepath() string {
	return c.filepath
}

type Cluster struct {
	Brokers []string `mapstructure:"brokers"`
	TLS     *TLS     `mapstructure:"tls"`
	SASL    *SASL    `mapstructure:"sasl"`
}

type TLS struct {
	Insecure bool   `mapstructure:"insecure"`
	CAFile   string `mapstructure:"cafile"`
}

type SASL struct {
	Mechanism string `mapstructure:"mechanism"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
}

func (c *Config) Cluster(profile string) ([]string, *sarama.Config, error) {
	if profile == "" {
		profile = c.DefaultCluster
	}

	prof, ok := c.Clusters[profile]
	if !ok {
		log.Fatal("No profile specified")
	}

	cfg := sarama.NewConfig()
	if prof.TLS != nil {
		cfg.Net.TLS.Enable = true

		raw, err := os.ReadFile(prof.TLS.CAFile)
		if err != nil {
			log.Fatal(err)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(raw)
		cfg.Net.TLS.Config = &tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: prof.TLS.Insecure,
		}
	}

	if prof.SASL != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLMechanism(prof.SASL.Mechanism)
		cfg.Net.SASL.User = prof.SASL.Username
		cfg.Net.SASL.Password = prof.SASL.Password
	}

	return prof.Brokers, cfg, nil
}

var defaultConfig = &Config{
	DefaultCluster: "default",
	Clusters: map[string]*Cluster{
		"default": {
			Brokers: []string{"127.0.0.1:9092"},
		},
	},
}

func LoadConfig(cfgFilepath string) (*Config, error) {
	cfgRoot := path.Join(xdg.ConfigHome, "kafka")
	if _, err := os.Stat(cfgRoot); os.IsNotExist(err) {
		if err := os.Mkdir(cfgRoot, 0755); err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}
	}

	cfg := &Config{
		filepath: viper.GetViper().ConfigFileUsed(),
	}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal(err)
	}

	return cfg, nil
}
