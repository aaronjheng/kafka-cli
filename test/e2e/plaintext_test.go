package e2e_test

import (
	"testing"
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

	messages := "hello-e2e-test\ne2e-produce-message\n"

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce messages failed: %v", err)
	}
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
