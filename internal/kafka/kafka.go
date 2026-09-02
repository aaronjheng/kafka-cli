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

	tunnel *ssh.Tunnel
}

// Close closes the client and the SSH tunnel it uses.
func (k *Kafka) Close() error {
	return errors.Join(k.Client.Close(), closeTunnel(k.tunnel))
}

var errInvalidTLSCA = errors.New("invalid TLS CA")

func newSaramaConfig(ctx context.Context, clusterConfig *Config) (*sarama.Config, *ssh.Tunnel, error) {
	err := clusterConfig.Validate()
	if err != nil {
		return nil, nil, fmt.Errorf("cluster config validation error: %w", err)
	}

	saramaCfg := sarama.NewConfig()

	err = configureTLS(saramaCfg, clusterConfig.TLS)
	if err != nil {
		return nil, nil, fmt.Errorf("configureTLS error: %w", err)
	}

	configureSASL(saramaCfg, clusterConfig.SASL)

	tunnel, err := configureSSH(ctx, saramaCfg, clusterConfig.SSH)
	if err != nil {
		return nil, nil, fmt.Errorf("configureSSH error: %w", err)
	}

	return saramaCfg, tunnel, nil
}

func configureTLS(saramaCfg *sarama.Config, tlsConfig *TLS) error {
	if tlsConfig == nil {
		return nil
	}

	saramaCfg.Net.TLS.Enable = true
	saramaCfg.Net.TLS.Config = &tls.Config{
		// #nosec G402 -- user-controlled option for self-signed/dev clusters.
		InsecureSkipVerify: tlsConfig.Insecure,
	}

	if tlsConfig.CAFile == "" {
		return nil
	}

	raw, err := os.ReadFile(tlsConfig.CAFile)
	if err != nil {
		return fmt.Errorf("os.ReadFile error: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(raw) {
		return fmt.Errorf("parse TLS CA file %q: %w", tlsConfig.CAFile, errInvalidTLSCA)
	}

	saramaCfg.Net.TLS.Config.RootCAs = certPool

	return nil
}

func configureSASL(saramaCfg *sarama.Config, saslConfig *SASL) {
	if saslConfig == nil {
		return
	}

	saramaCfg.Net.SASL.Enable = true
	saramaCfg.Net.SASL.Mechanism = sarama.SASLMechanism(saslConfig.Mechanism)
	saramaCfg.Net.SASL.User = saslConfig.Username
	saramaCfg.Net.SASL.Password = saslConfig.Password

	switch strings.ToUpper(saslConfig.Mechanism) {
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

func configureSSH(ctx context.Context, saramaCfg *sarama.Config, sshConfig *ssh.Config) (*ssh.Tunnel, error) {
	if sshConfig == nil {
		// A no-op tunnel keeps cleanup uniform for clusters without SSH.
		return &ssh.Tunnel{}, nil
	}

	tunnel, err := ssh.NewTunnel(ctx, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh.NewTunnel error: %w", err)
	}

	saramaCfg.Net.Proxy.Enable = true
	saramaCfg.Net.Proxy.Dialer = tunnel

	return tunnel, nil
}

func closeTunnel(tunnel *ssh.Tunnel) error {
	if tunnel == nil {
		return nil
	}

	err := tunnel.Close()
	if err != nil {
		return fmt.Errorf("tunnel.Close error: %w", err)
	}

	return nil
}

func New(ctx context.Context, clusterConfig *Config) (*Kafka, error) {
	saramaCfg, tunnel, err := newSaramaConfig(ctx, clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	client, err := sarama.NewClient(clusterConfig.Brokers, saramaCfg)
	if err != nil {
		_ = closeTunnel(tunnel)

		return nil, fmt.Errorf("sarama.NewClient error: %w", err)
	}

	return &Kafka{
		Client: client,
		tunnel: tunnel,
	}, nil
}

// SyncProducer wraps a sarama sync producer together with its SSH tunnel.
type SyncProducer struct {
	sarama.SyncProducer

	tunnel *ssh.Tunnel
}

func NewSyncProducer(ctx context.Context, cfg *Config) (*SyncProducer, error) {
	saramaCfg, tunnel, err := newSaramaConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	saramaCfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		_ = closeTunnel(tunnel)

		return nil, fmt.Errorf("sarama.NewSyncProducer error: %w", err)
	}

	return &SyncProducer{
		SyncProducer: producer,
		tunnel:       tunnel,
	}, nil
}

// Close closes the producer and the SSH tunnel it uses.
func (p *SyncProducer) Close() error {
	return errors.Join(p.SyncProducer.Close(), closeTunnel(p.tunnel))
}

type PartitionReader struct {
	consumer          sarama.Consumer
	partitionConsumer sarama.PartitionConsumer
	tunnel            *ssh.Tunnel
}

func NewPartitionReader(
	ctx context.Context,
	cfg *Config,
	topic string,
	partition int32,
	offset int64,
) (*PartitionReader, error) {
	saramaCfg, tunnel, err := newSaramaConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("newSaramaConfig error: %w", err)
	}

	saramaCfg.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Brokers, saramaCfg)
	if err != nil {
		_ = closeTunnel(tunnel)

		return nil, fmt.Errorf("sarama.NewConsumer error: %w", err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		_ = consumer.Close()
		_ = closeTunnel(tunnel)

		return nil, fmt.Errorf("consumer.ConsumePartition error: %w", err)
	}

	return &PartitionReader{
		consumer:          consumer,
		partitionConsumer: partitionConsumer,
		tunnel:            tunnel,
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

	err = closeTunnel(r.tunnel)
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
