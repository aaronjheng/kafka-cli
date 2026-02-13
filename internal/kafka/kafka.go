package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/aaronjheng/kafka-cli/internal/ssh"
)

type Kafka struct {
	sarama.Client
}

func New(c *Config) (*Kafka, error) {
	cfg := sarama.NewConfig()
	if c.TLS != nil {
		cfg.Net.TLS.Enable = true

		raw, err := os.ReadFile(c.TLS.CAFile)
		if err != nil {
			return nil, fmt.Errorf("os.ReadFile error: %w", err)
		}

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(raw)
		cfg.Net.TLS.Config = &tls.Config{
			RootCAs: certPool,
			// #nosec G402 -- user-controlled option for self-signed/dev clusters.
			InsecureSkipVerify: c.TLS.Insecure,
		}
	}

	if c.SASL != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLMechanism(c.SASL.Mechanism)
		cfg.Net.SASL.User = c.SASL.Username
		cfg.Net.SASL.Password = c.SASL.Password
	}

	if c.SSH != nil {
		dialer, err := ssh.NewProxyDialer(c.SSH)
		if err != nil {
			return nil, fmt.Errorf("newSSHDialFunc error: %w", err)
		}

		cfg.Net.Proxy.Enable = true
		cfg.Net.Proxy.Dialer = dialer
		cfg.Net.DialTimeout = 0
	}

	client, err := sarama.NewClient(c.Brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewClient error: %w", err)
	}

	return &Kafka{
		Client: client,
	}, nil
}

func NewWriter(cfg *Config, topic string) (*kafkago.Writer, error) {
	dialer, err := NewDialer(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewDialer error: %w", err)
	}

	return kafkago.NewWriter(kafkago.WriterConfig{
		Brokers: cfg.Brokers,
		Topic:   topic,
		Dialer:  dialer,
	}), nil
}

func NewPartitionReader(
	brokers []string,
	dialer *kafkago.Dialer,
	topic string,
	partition int32,
) (*kafkago.Reader, error) {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: int(partition),
		Dialer:    dialer,
	}), nil
}
