package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	Host         string `mapstructure:"host"`
	Port         int32  `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	IdentityFile string `mapstructure:"identity_file"`
}

type Client struct {
	*ssh.Client
}

func NewClient(cfg *Config) (*Client, error) {
	identityFiles := []string{
		"~/.ssh/id_ed25519",
		"~/.ssh/id_ecdsa",
		"~/.ssh/id_dsa",
		"~/.ssh/id_rsa",
	}

	if cfg.IdentityFile != "" {
		identityFiles = append([]string{cfg.IdentityFile}, identityFiles...)
	}

	signers, err := sshSignersFromIdentityFiles(identityFiles)
	if err != nil {
		return nil, fmt.Errorf("sshSignersFromIdentityFiles error: %w", err)
	}

	sshCfg := &ssh.ClientConfig{
		User: cfg.Username,
		// #nosec G106 -- compatibility mode: allow connecting to hosts not present in known_hosts.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial error: %w", err)
	}

	client := &Client{
		Client: sshClient,
	}

	return client, nil
}

func (c *Client) Dial(ctx context.Context, protocol, address string) (net.Conn, error) {
	conn, err := c.Client.DialContext(ctx, protocol, address)
	if err != nil {
		return nil, fmt.Errorf("ssh.Client.DialContext error: %w", err)
	}

	return conn, nil
}

// Deprecated: Use Dial instead.
func (c *Client) DialContext(ctx context.Context, protocol, address string) (net.Conn, error) {
	return c.Dial(ctx, protocol, address)
}

func sshSignersFromIdentityFiles(identityFiles []string) ([]ssh.Signer, error) {
	signers := make([]ssh.Signer, 0, len(identityFiles))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("os.UserHomeDir error: %w", err)
	}

	for _, f := range identityFiles {
		if homeDir != "" && strings.HasPrefix(f, "~/") {
			f = filepath.Join(homeDir, f[2:])
		}

		_, err := os.Stat(f)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("os.Stat error: %w", err)
		}

		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("os.ReadFile error: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(raw)
		if err != nil {
			return nil, fmt.Errorf("ssh.ParsePrivateKey error: %w", err)
		}

		signers = append(signers, signer)
	}

	return signers, nil
}
