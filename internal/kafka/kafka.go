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

func New(clusterConfig *Config) (*Kafka, error) {
	saramaCfg := sarama.NewConfig()
	if clusterConfig.TLS != nil {
		saramaCfg.Net.TLS.Enable = true

		raw, err := os.ReadFile(clusterConfig.TLS.CAFile)
		if err != nil {
			return nil, fmt.Errorf("os.ReadFile error: %w", err)
		}

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(raw)
		saramaCfg.Net.TLS.Config = &tls.Config{
			RootCAs: certPool,
			// #nosec G402 -- user-controlled option for self-signed/dev clusters.
			InsecureSkipVerify: clusterConfig.TLS.Insecure,
		}
	}

	if clusterConfig.SASL != nil {
		saramaCfg.Net.SASL.Enable = true
		saramaCfg.Net.SASL.Mechanism = sarama.SASLMechanism(clusterConfig.SASL.Mechanism)
		saramaCfg.Net.SASL.User = clusterConfig.SASL.Username
		saramaCfg.Net.SASL.Password = clusterConfig.SASL.Password
	}

	if clusterConfig.SSH != nil {
		dialer, err := ssh.NewDialerFunc(clusterConfig.SSH)
		if err != nil {
			return nil, fmt.Errorf("ssh.NewDialerFunc error: %w", err)
		}

		saramaCfg.Net.Proxy.Enable = true
		saramaCfg.Net.Proxy.Dialer = dialer
	}

	client, err := sarama.NewClient(clusterConfig.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewClient error: %w", err)
	}

	return &Kafka{
		Client: client,
	}, nil
}

func NewWriter(cfg *Config, topic string) (*kafkago.Writer, error) {
	transport := &kafkago.Transport{}

	if cfg.SSH != nil {
		dialerFunc, err := ssh.NewDialerFunc(cfg.SSH)
		if err != nil {
			return nil, fmt.Errorf("ssh.NewDialerFunc error: %w", err)
		}

		transport.Dial = dialerFunc
	}

	if cfg.TLS != nil {
		tlsConfig, err := newTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("newTLSConfig error: %w", err)
		}

		transport.TLS = tlsConfig
	}

	if cfg.SASL != nil {
		mechanism, err := newSASLMechanism(cfg.SASL)
		if err != nil {
			return nil, fmt.Errorf("newSASLMechanism error: %w", err)
		}

		transport.SASL = mechanism
	}

	return &kafkago.Writer{
		Addr:      kafkago.TCP(cfg.Brokers...),
		Topic:     topic,
		Transport: transport,
	}, nil
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
