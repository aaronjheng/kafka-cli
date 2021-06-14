package config

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"path"

	"github.com/Shopify/sarama"
	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml"
)

type TLS struct {
	Insecure bool   `toml:"insecure"`
	CAFile   string `toml:"cafile"`
}

type SASL struct {
	Mechanism string `toml:"mechanism"`
	Username  string `toml:"username"`
	Password  string `toml:"password"`
}

type Cluster struct {
	Brokers []string `toml:"brokers"`
	TLS     *TLS     `toml:"tls"`
	SASL    *SASL    `toml:"sasl"`
}

type Config struct {
	DefaultCluster string              `toml:"default_cluster"`
	Clusters       map[string]*Cluster `toml:"clusters"`
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

func LoadConfig() (*Config, error) {
	cfgRoot := path.Join(xdg.ConfigHome, "kafka")
	if _, err := os.Stat(cfgRoot); os.IsNotExist(err) {
		if err := os.Mkdir(cfgRoot, 0755); err != nil {
			log.Fatal(err)
		}
	}

	isConfigFileExists := true
	configFilepath := path.Join(cfgRoot, "config.toml2")
	if _, err := os.Stat(configFilepath); err != nil {
		if os.IsNotExist(err) {
			isConfigFileExists = false
		} else {
			log.Fatal(err)
		}
	}

	configFile, err := os.OpenFile(configFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if !isConfigFileExists {
		if _, err := configFile.WriteString("# Kafkactl configuration\n\n"); err != nil {
			log.Fatal(err)
		}
	}

	defer configFile.Close()

	cfg := &Config{}
	if err := toml.NewDecoder(configFile).Decode(cfg); err != nil {
		log.Fatal(err)
	}

	return cfg, nil
}
