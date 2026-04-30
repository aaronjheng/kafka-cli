package e2e_test

import (
	"os"
	"testing"
)

const (
	tlsTestBroker = "127.0.0.1:9095"
	tlsCACertEnv  = "TLS_TEST_CA_CERT"
)

func requireTLS(t *testing.T) {
	t.Helper()

	if os.Getenv(tlsCACertEnv) == "" {
		t.Skip("TLS_TEST_CA_CERT not set, skipping TLS E2E test")
	}
}

func tlsConfig() string {
	caCertPath := os.Getenv(tlsCACertEnv)

	return `
default_cluster: tls

clusters:
  tls:
    brokers:
      - ` + tlsTestBroker + `
    tls:
      insecure: true
      cafile: ` + caCertPath + `
`
}

func TestClusterDescribe_TLS(t *testing.T) {
	t.Parallel()
	requireTLS(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, tlsConfig())

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe failed: %v", err)
	}

	if !StringsContains(output, "Cluster ID:") {
		t.Errorf("expected 'Cluster ID:' in output, got: %s", output)
	}

	if !StringsContains(output, "Brokers") {
		t.Errorf("expected 'Brokers' in output, got: %s", output)
	}
}

func TestTopicCRUD_TLS(t *testing.T) {
	t.Parallel()
	requireTLS(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, tlsConfig())

	WaitForKafka(t, cli)

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
}

func TestTopicProduceConsume_TLS(t *testing.T) {
	t.Parallel()
	requireTLS(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, tlsConfig())

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	expectedMessages := []string{"hello-tls-e2e-test", "tls-produce-message"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
}

func TestGroupList_TLS(t *testing.T) {
	t.Parallel()
	requireTLS(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, tlsConfig())

	WaitForKafka(t, cli)

	_, err := cli.Run(t.Context(), "group", "list")
	if err != nil {
		t.Fatalf("list consumer groups failed: %v", err)
	}
}
