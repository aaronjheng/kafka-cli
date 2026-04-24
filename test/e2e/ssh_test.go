package e2e_test

import (
	"os"
	"testing"
)

const (
	sshKeyPathEnv = "SSH_TEST_KEY_PATH"
	sshTestUser   = "kafkauser"
	sshTestPort   = "2222"
	sshTestHost   = "127.0.0.1"
)

var sshKeyPath string //nolint:gochecknoglobals // test-only: set by TestMain

func TestMain(m *testing.M) {
	sshKeyPath = os.Getenv(sshKeyPathEnv)

	os.Exit(m.Run())
}

func requireSSH(t *testing.T) {
	t.Helper()

	if sshKeyPath == "" {
		t.Skip("SSH_TEST_KEY_PATH not set, skipping SSH E2E test")
	}
}

func sshConfig() string {
	return `
default_cluster: ssh

clusters:
  ssh:
    brokers:
      - 127.0.0.1:9092
    ssh:
      host: ` + sshTestHost + `
      port: ` + sshTestPort + `
      user: ` + sshTestUser + `
      identity_file: ` + sshKeyPath + `
`
}

func TestClusterDescribe_SSH(t *testing.T) {
	t.Parallel()
	requireSSH(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, sshConfig())

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe failed: %v", err)
	}

	if !StringsContains(output, "Controller ID:") {
		t.Errorf("expected 'Controller ID:' in output, got: %s", output)
	}

	if !StringsContains(output, "Brokers:") {
		t.Errorf("expected 'Brokers:' in output, got: %s", output)
	}

	if !StringsContains(output, "Topics:") {
		t.Errorf("expected 'Topics:' in output, got: %s", output)
	}
}

func TestTopicProduceConsume_SSH(t *testing.T) {
	t.Parallel()
	requireSSH(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, sshConfig())

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	expectedMessages := []string{"hello-ssh-e2e-test", "second-message"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages)
}

func TestConnectionFailure_SSH(t *testing.T) {
	t.Parallel()
	requireSSH(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: ssh

clusters:
  ssh:
    brokers:
      - 127.0.0.1:9092
    ssh:
      host: 127.0.0.1
      port: 2222
      user: kafkauser
      identity_file: /nonexistent/key
`)

	_, err := cli.Run(t.Context(), "cluster", "describe")
	if err == nil {
		t.Error("expected error when using nonexistent SSH key")
	}
}
