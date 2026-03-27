package ssh

import (
	"golang.org/x/net/proxy"
)

func NewProxyDialer(cfg *Config) (proxy.Dialer, error) {
	dialerFunc, err := newDialerFunc(cfg)
	if err != nil {
		return nil, err
	}

	return dialerFunc, nil
}
