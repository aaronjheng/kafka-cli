package e2e_test

import (
	"testing"
)

func TestTopicProduce_KeySeparator(t *testing.T) {
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

	messages := "key1:value1\nkey2:value2\nvalue-only\n"

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic, "--key-separator", ":")
	if err != nil {
		t.Fatalf("produce with key-separator failed: %v", err)
	}

	expectedMessages := []string{"key1:value1", "key2:value2", "value-only"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
}

func TestTopicProduce_EmptyLinesSkipped(t *testing.T) {
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

	messages := "msg1\n\n\nmsg2\n"

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce with empty lines failed: %v", err)
	}

	expectedMessages := []string{"msg1", "msg2"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
}

func TestTopicProduceConsume_AllPartitions(t *testing.T) {
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

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	expectedMessages := []string{"hello-all-partitions", "multi-partition-msg"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages)
}

func TestTopicCreate_DefaultFlags(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "create", topic)
	if err != nil {
		t.Fatalf("create topic with default flags failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	describeOutput, err := cli.Run(t.Context(), "topic", "describe", topic)
	if err != nil {
		t.Fatalf("describe topic failed: %v", err)
	}

	partitions := ExtractPartitionCount(describeOutput)
	if partitions != "3" {
		t.Errorf("expected 3 partitions (default), got %q in output: %s", partitions, describeOutput)
	}
}

func TestTopicDelete_MultipleTopics(t *testing.T) {
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

	topic1 := UniqueTopicName(t) + "-a"
	topic2 := UniqueTopicName(t) + "-b"

	_, err := cli.Run(t.Context(), "topic", "create", topic1, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic1 failed: %v", err)
	}

	_, err = cli.Run(t.Context(), "topic", "create", topic2, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic2 failed: %v", err)
	}

	_, err = cli.Run(t.Context(), "topic", "delete", topic1, topic2)
	if err != nil {
		t.Fatalf("delete multiple topics failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "list")
	if err != nil {
		t.Fatalf("list topics after delete failed: %v", err)
	}

	if TopicExistsInOutput(output, topic1) {
		t.Errorf("topic %q should have been deleted", topic1)
	}

	if TopicExistsInOutput(output, topic2) {
		t.Errorf("topic %q should have been deleted", topic2)
	}
}

func TestGroupDelete_MultipleGroups(t *testing.T) {
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
	group1 := UniqueGroupName(t) + "-a"
	group2 := UniqueGroupName(t) + "-b"

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "1", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	ProduceAndConsumeWithGroup(t, cli, topic, group1, []string{"hello-multi-group-1"})
	ProduceAndConsumeWithGroup(t, cli, topic, group2, []string{"hello-multi-group-2"})

	_, err = cli.Run(t.Context(), "group", "delete", group1, group2)
	if err != nil {
		t.Fatalf("delete multiple groups failed: %v", err)
	}
}

func TestTopicProduce_KeySeparator_LineWithoutSeparator(t *testing.T) {
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

	messages := "key1:value1\nno-separator-here\nkey2:value2\n"

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic, "--key-separator", ":")
	if err != nil {
		t.Fatalf("produce with key-separator and line without separator failed: %v", err)
	}

	expectedMessages := []string{"key1:value1", "no-separator-here", "key2:value2"}

	ProduceAndAssertConsumed(t, cli, topic, expectedMessages, "-p", "0")
}

func TestTopicGetOffsets_AfterProduce(t *testing.T) {
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

	_, err = cli.RunWithStdin(t.Context(), "msg1\nmsg2\nmsg3\n", "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce messages failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "get-offsets", topic)
	if err != nil {
		t.Fatalf("get-offsets failed: %v", err)
	}

	if !StringsContains(output, "3") {
		t.Errorf("expected newest offset 3 in output, got: %s", output)
	}
}
