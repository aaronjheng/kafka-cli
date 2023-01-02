package config

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/Shopify/sarama"
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
