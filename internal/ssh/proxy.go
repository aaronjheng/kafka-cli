package ssh

import (
	"golang.org/x/net/proxy"
)

//nolint:ireturn // Return interface is required by the proxy package.
func NewProxyDialer(cfg *Config) (proxy.Dialer, error) {
	dialerFunc, err := newDialerFunc(cfg)
	if err != nil {
		return nil, err
	}

	return dialerFunc, nil
}
