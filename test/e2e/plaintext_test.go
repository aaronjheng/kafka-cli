package e2e_test

import (
	"fmt"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

func TestClusterDescribe_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

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

func TestTopicCRUD_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

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

func TestTopicProduceConsume_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	expectedMessages := []string{"hello-e2e-test", "e2e-produce-message"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
}

func TestGroupList_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

	_, err := cli.Run(t.Context(), "group", "list")
	if err != nil {
		t.Fatalf("list consumer groups failed: %v", err)
	}
}

func assertOffsetEquals(t *testing.T, client *kafkago.Client, group, topic string, expected int64) {
	t.Helper()

	offset := fetchGroupOffset(t, client, group, topic, 0)
	if offset != expected {
		t.Fatalf("expected offset %d, got %d", expected, offset)
	}
}

func testGroupResetOffsets(t *testing.T, cli *KafkaCLI, brokerAddr, saslMech, username, password string) {
	t.Helper()

	topic := UniqueTopicName(t)
	group := fmt.Sprintf("e2e-group-%d", time.Now().UnixMilli())

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	messages := "msg1\nmsg2\nmsg3\n"

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce messages failed: %v", err)
	}

	client, err := newKafkaGoClient(brokerAddr, saslMech, username, password)
	if err != nil {
		t.Fatalf("failed to create kafka client: %v", err)
	}

	setGroupOffset(t, client, group, topic, 0, 3)
	assertOffsetEquals(t, client, group, topic, 3)

	_, err = cli.Run(t.Context(), "group", "reset-offsets", group, "--topic", topic, "--to-earliest")
	if err != nil {
		t.Fatalf("reset-offsets to-earliest failed: %v", err)
	}

	assertOffsetEquals(t, client, group, topic, 0)

	_, err = cli.Run(t.Context(), "group", "reset-offsets", group, "--topic", topic, "--to-latest")
	if err != nil {
		t.Fatalf("reset-offsets to-latest failed: %v", err)
	}

	assertOffsetEquals(t, client, group, topic, 3)

	_, err = cli.Run(t.Context(), "group", "reset-offsets", group, "--topic", topic, "--to-offset", "1")
	if err != nil {
		t.Fatalf("reset-offsets to-offset failed: %v", err)
	}

	assertOffsetEquals(t, client, group, topic, 1)
}

func TestGroupResetOffsets_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

	testGroupResetOffsets(t, cli, "127.0.0.1:9092", "", "", "")
}

func TestGroupResetOffsets_Validation(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

	group := fmt.Sprintf("e2e-group-%d", time.Now().UnixMilli())

	// Missing topic.
	_, err := cli.Run(t.Context(), "group", "reset-offsets", group, "--to-earliest")
	if err == nil {
		t.Fatal("expected error when topic is missing")
	}

	// Missing strategy.
	_, err = cli.Run(t.Context(), "group", "reset-offsets", group, "--topic", "test-topic")
	if err == nil {
		t.Fatal("expected error when reset strategy is missing")
	}

	// Multiple strategies.
	_, err = cli.Run(t.Context(), "group", "reset-offsets", group, "--topic", "test-topic", "--to-earliest", "--to-latest")
	if err == nil {
		t.Fatal("expected error when multiple reset strategies are specified")
	}
}
