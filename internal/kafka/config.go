package kafka

import (
	"github.com/aaronjheng/kafka-cli/internal/ssh"
)

type Config struct {
	Brokers []string    `mapstructure:"brokers"`
	TLS     *TLS        `mapstructure:"tls"`
	SASL    *SASL       `mapstructure:"sasl"`
	SSH     *ssh.Config `mapstructure:"ssh_tunnel"`
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
