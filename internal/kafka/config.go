package kafka

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aaronjheng/kafka-cli/internal/ssh"
)

var (
	ErrConfigRequired = errors.New("kafka config is required")
	errBrokerRequired = errors.New("broker is required")
)

type Config struct {
	Brokers []string    `mapstructure:"brokers"`
	TLS     *TLS        `mapstructure:"tls"`
	SASL    *SASL       `mapstructure:"sasl"`
	SSH     *ssh.Config `mapstructure:"ssh"`
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

func (c *Config) Validate() error {
	if c == nil {
		return ErrConfigRequired
	}

	if len(c.Brokers) == 0 {
		return errBrokerRequired
	}

	for i, broker := range c.Brokers {
		if strings.TrimSpace(broker) == "" {
			return fmt.Errorf("%w at index %d", errBrokerRequired, i)
		}
	}

	return nil
}
