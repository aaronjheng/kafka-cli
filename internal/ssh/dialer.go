package ssh

import (
	"context"
	"fmt"
	"net"
)

type DialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func NewDialerFunc(cfg *Config) (DialerFunc, error) {
	return newDialerFunc(cfg)
}

func newDialerFunc(cfg *Config) (DialerFunc, error) {
	sshClient, err := NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewClient error: %w", err)
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return sshClient.Dial(ctx, network, addr)
	}, nil
}

func (d DialerFunc) Dial(network, addr string) (net.Conn, error) {
	return d(context.Background(), network, addr)
}

func (d DialerFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d(ctx, network, addr)
}
