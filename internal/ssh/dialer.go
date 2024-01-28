package ssh

import (
	"context"
	"net"
)

type DialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func (d DialerFunc) Dial(network, addr string) (c net.Conn, err error) {
	return d(context.Background(), network, addr)
}

func (d DialerFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d(ctx, network, addr)
}

func NewDialerFunc(cfg *Config) (DialerFunc, error) {
	return newDialerFunc(cfg)
}

func newDialerFunc(cfg *Config) (DialerFunc, error) {
	sshClient, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return sshClient.Dial(ctx, network, addr)
	}, nil
}
