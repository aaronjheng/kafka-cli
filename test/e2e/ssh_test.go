package e2e_test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestTopicList_SSHProxy(t *testing.T) {
	t.Parallel()

	sshKeyPath := filepath.Join(".", "ssh_key")

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, fmt.Sprintf(`
default_cluster: ssh-proxy

clusters:
  ssh-proxy:
    brokers:
      - 127.0.0.1:9092
    ssh:
      host: 127.0.0.1
      port: 2222
      user: testuser
      identity_file: %s
`, sshKeyPath))

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic via SSH proxy failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "list")
	if err != nil {
		t.Fatalf("list topics via SSH proxy failed: %v", err)
	}

	if !TopicExistsInOutput(output, topic) {
		t.Errorf("topic %q not found in SSH proxy list output", topic)
	}

	_, err = cli.Run(t.Context(), "topic", "delete", topic)
	if err != nil {
		t.Fatalf("delete topic via SSH proxy failed: %v", err)
	}
}

func TestClusterDescribe_SSHProxy(t *testing.T) {
	t.Parallel()

	sshKeyPath := filepath.Join(".", "ssh_key")

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, fmt.Sprintf(`
default_cluster: ssh-proxy

clusters:
  ssh-proxy:
    brokers:
      - 127.0.0.1:9092
    ssh:
      host: 127.0.0.1
      port: 2222
      user: testuser
      identity_file: %s
`, sshKeyPath))

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe via SSH proxy failed: %v", err)
	}

	if !StringsContains(output, "Controller ID:") {
		t.Errorf("expected 'Controller ID:' in SSH proxy cluster describe output, got: %s", output)
	}
}
