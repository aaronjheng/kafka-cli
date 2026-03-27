package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"

	"github.com/aaronjheng/kafka-cli/internal/ssh"
)

const defaultDialTimeout = 10 * time.Second

var errUnsupportedSASLMechanism = errors.New("unsupported sasl mechanism")

func NewDialer(cfg *Config) (*kafkago.Dialer, error) {
	dialer := &kafkago.Dialer{
		Timeout: defaultDialTimeout,
	}

	if cfg.TLS != nil {
		tlsConfig, err := newTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("newTLSConfig error: %w", err)
		}

		dialer.TLS = tlsConfig
	}

	if cfg.SASL != nil {
		mechanism, err := newSASLMechanism(cfg.SASL)
		if err != nil {
			return nil, fmt.Errorf("newSASLMechanism error: %w", err)
		}

		dialer.SASLMechanism = mechanism
	}

	if cfg.SSH != nil {
		dialerFunc, err := ssh.NewDialerFunc(cfg.SSH)
		if err != nil {
			return nil, fmt.Errorf("ssh.NewDialerFunc error: %w", err)
		}

		dialer.DialFunc = dialerFunc
	}

	return dialer, nil
}

func newTLSConfig(cfg *TLS) (*tls.Config, error) {
	raw, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile error: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(raw)

	return &tls.Config{
		RootCAs: certPool,
		// #nosec G402 -- user-controlled option for self-signed/dev clusters.
		InsecureSkipVerify: cfg.Insecure,
	}, nil
}

func newSASLMechanism(cfg *SASL) (sasl.Mechanism, error) {
	switch strings.ToUpper(cfg.Mechanism) {
	case "PLAIN":
		return plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}, nil
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("scram.Mechanism error: %w", err)
		}

		return mechanism, nil
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("scram.Mechanism error: %w", err)
		}

		return mechanism, nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedSASLMechanism, cfg.Mechanism)
	}
}
