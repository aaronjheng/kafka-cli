package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"

	"github.com/aaronjheng/kafka-cli/internal/ssh"
)

type Kafka struct {
	sarama.Client
}

func newSaramaConfig(clusterConfig *Config) (*sarama.Config, error) {
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

		switch strings.ToUpper(clusterConfig.SASL.Mechanism) {
		case "SCRAM-SHA-256":
			saramaCfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return newSaramaSCRAMClient(scram.SHA256)
			}
		case "SCRAM-SHA-512":
			saramaCfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return newSaramaSCRAMClient(scram.SHA512)
			}
		}
	}

	if clusterConfig.SSH != nil {
		dialer, err := ssh.NewDialerFunc(clusterConfig.SSH)
		if err != nil {
			return nil, fmt.Errorf("ssh.NewDialerFunc error: %w", err)
		}

		saramaCfg.Net.Proxy.Enable = true
		saramaCfg.Net.Proxy.Dialer = dialer
	}

	return saramaCfg, nil
}

func New(clusterConfig *Config) (*Kafka, error) {
	saramaCfg, err := newSaramaConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	client, err := sarama.NewClient(clusterConfig.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewClient error: %w", err)
	}

	return &Kafka{
		Client: client,
	}, nil
}

func NewSyncProducer(cfg *Config) (sarama.SyncProducer, error) {
	saramaCfg, err := newSaramaConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	saramaCfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewSyncProducer error: %w", err)
	}

	return producer, nil
}

type PartitionReader struct {
	consumer          sarama.Consumer
	partitionConsumer sarama.PartitionConsumer
}

func NewPartitionReader(cfg *Config, topic string, partition int32, offset int64) (*PartitionReader, error) {
	saramaCfg, err := newSaramaConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	saramaCfg.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewConsumer error: %w", err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		_ = consumer.Close()

		return nil, fmt.Errorf("consumer.ConsumePartition error: %w", err)
	}

	return &PartitionReader{
		consumer:          consumer,
		partitionConsumer: partitionConsumer,
	}, nil
}

func (r *PartitionReader) ReadMessage(ctx context.Context) (*sarama.ConsumerMessage, error) {
	select {
	case msg, ok := <-r.partitionConsumer.Messages():
		if !ok {
			return nil, io.EOF
		}

		return msg, nil
	case err, ok := <-r.partitionConsumer.Errors():
		if !ok {
			return nil, io.EOF
		}

		return nil, err.Err
	case <-ctx.Done():
		return nil, fmt.Errorf("context error: %w", ctx.Err())
	}
}

func (r *PartitionReader) Close() error {
	var errs []error

	err := r.partitionConsumer.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err = r.consumer.Close()
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
