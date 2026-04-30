package e2e_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

const (
	saslTestBroker = "127.0.0.1:9094"
	saslUsername   = "client"
	saslPassword   = "client-secret"

	scram256Username = "scram256-client"
	scram256Password = "scram256-secret"

	scram512Username = "scram512-client"
	scram512Password = "scram512-secret"
)

var scramUsersOnce sync.Once //nolint:gochecknoglobals // test-only: ensures SCRAM users are created once

var errSCRAMUsersCreate error

func ensureSCRAMUsers(t *testing.T) {
	t.Helper()

	scramUsersOnce.Do(func() {
		errSCRAMUsersCreate = createSCRAMUsers()
	})

	if errSCRAMUsersCreate != nil {
		t.Fatalf("failed to create SCRAM users: %v", errSCRAMUsersCreate)
	}
}

func createSCRAMUsers() error {
	kafkaContainer := os.Getenv("KAFKA_CONTAINER_NAME")
	if kafkaContainer == "" {
		kafkaContainer = "kafka-ssh-kafka-ssh-test"
	}

	creds := []struct {
		username  string
		mechanism string
		password  string
	}{
		{scram256Username, "SCRAM-SHA-256", scram256Password},
		{scram512Username, "SCRAM-SHA-512", scram512Password},
	}

	for _, cred := range creds {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		args := []string{
			"exec", kafkaContainer,
			"/opt/kafka/bin/kafka-configs.sh",
			"--bootstrap-server", "127.0.0.1:9092",
			"--alter",
			"--add-config", fmt.Sprintf("%s=[password=%s]", cred.mechanism, cred.password),
			"--entity-type", "users",
			"--entity-name", cred.username,
		}

		cmd := exec.CommandContext(ctx, "podman", args...) //nolint:gosec // test-only: controlled args
		output, err := cmd.CombinedOutput()

		cancel()

		if err != nil {
			return fmt.Errorf("kafka-configs.sh for %s: %w, output: %s", cred.username, err, output)
		}
	}

	return nil
}

type saslMechanismConfig struct {
	mechanism string
	username  string
	password  string
	setup     func(t *testing.T)
}

var saslMechanisms = map[string]saslMechanismConfig{ //nolint:gochecknoglobals // test-only: static test config
	"PLAIN": {
		mechanism: "PLAIN",
		username:  saslUsername,
		password:  saslPassword,
		setup:     nil,
	},
	"SCRAM-SHA-256": {
		mechanism: "SCRAM-SHA-256",
		username:  scram256Username,
		password:  scram256Password,
		setup:     ensureSCRAMUsers,
	},
	"SCRAM-SHA-512": {
		mechanism: "SCRAM-SHA-512",
		username:  scram512Username,
		password:  scram512Password,
		setup:     ensureSCRAMUsers,
	},
}

func saslConfig(mechanism, username, password string) string {
	return `
default_cluster: sasl

clusters:
  sasl:
    brokers:
      - ` + saslTestBroker + `
    sasl:
      mechanism: ` + mechanism + `
      username: ` + username + `
      password: ` + password + `
`
}

func newSASLCLI(t *testing.T, mechConfig saslMechanismConfig) *KafkaCLI {
	t.Helper()

	if mechConfig.setup != nil {
		mechConfig.setup(t)
	}

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, saslConfig(mechConfig.mechanism, mechConfig.username, mechConfig.password))

	WaitForKafka(t, cli)

	return cli
}

func TestClusterDescribe_SASL(t *testing.T) {
	t.Parallel()

	for name, mechConfig := range saslMechanisms {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli := newSASLCLI(t, mechConfig)

			output, err := cli.Run(t.Context(), "cluster", "describe")
			if err != nil {
				t.Fatalf("cluster describe failed: %v", err)
			}

			assertClusterDescribeOutput(t, output)
		})
	}
}

func TestTopicCRUD_SASL(t *testing.T) {
	t.Parallel()

	for name, mechConfig := range saslMechanisms {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli := newSASLCLI(t, mechConfig)

			topic := UniqueTopicName(t)

			_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "3", "--replication-factor", "1")
			if err != nil {
				t.Fatalf("create topic failed: %v", err)
			}

			output, err := cli.Run(t.Context(), "topic", "list")
			if err != nil {
				t.Fatalf("list topics failed: %v", err)
			}

			if !TopicExistsInOutput(output, topic) {
				t.Errorf("topic %q not found in list output: %s", topic, output)
			}

			describeOutput, err := cli.Run(t.Context(), "topic", "describe", topic)
			if err != nil {
				t.Fatalf("describe topic failed: %v", err)
			}

			partitions := ExtractPartitionCount(describeOutput)
			if partitions != "3" {
				t.Errorf("expected 3 partitions, got %q in output: %s", partitions, describeOutput)
			}

			_, err = cli.Run(t.Context(), "topic", "delete", topic)
			if err != nil {
				t.Fatalf("delete topic failed: %v", err)
			}

			output, err = cli.Run(t.Context(), "topic", "list")
			if err != nil {
				t.Fatalf("list topics after delete failed: %v", err)
			}

			if TopicExistsInOutput(output, topic) {
				t.Errorf("topic %q should have been deleted but still appears in list", topic)
			}
		})
	}
}

func TestTopicProduceConsume_SASL(t *testing.T) {
	t.Parallel()

	for name, mechConfig := range saslMechanisms {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli := newSASLCLI(t, mechConfig)

			topic := UniqueTopicName(t)

			_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
			if err != nil {
				t.Fatalf("create topic failed: %v", err)
			}

			t.Cleanup(func() {
				_, _ = cli.Run(t.Context(), "topic", "delete", topic)
			})

			expectedMessages := []string{
				"hello-sasl-" + mechConfig.mechanism + "-e2e-test",
				mechConfig.mechanism + "-produce-message",
			}

			ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
		})
	}
}

func TestGroupList_SASL(t *testing.T) {
	t.Parallel()

	for name, mechConfig := range saslMechanisms {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli := newSASLCLI(t, mechConfig)

			_, err := cli.Run(t.Context(), "group", "list")
			if err != nil {
				t.Fatalf("list consumer groups failed: %v", err)
			}
		})
	}
}

func TestSASL_InvalidCredentials(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, saslConfig("PLAIN", "wrong-user", "wrong-password"))

	_, err := cli.Run(t.Context(), "cluster", "describe")
	if err == nil {
		t.Error("expected error when using invalid SASL credentials")
	}
}

func TestSASL_UnsupportedMechanism(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, saslConfig("UNSUPPORTED", "user", "pass"))

	_, err := cli.Run(t.Context(), "cluster", "describe")
	if err == nil {
		t.Error("expected error when using unsupported SASL mechanism")
	}
}

func assertClusterDescribeOutput(t *testing.T, output string) {
	t.Helper()

	if !StringsContains(output, "Cluster ID:") {
		t.Errorf("expected 'Cluster ID:' in output, got: %s", output)
	}

	if !StringsContains(output, "Brokers") {
		t.Errorf("expected 'Brokers' in output, got: %s", output)
	}
}
