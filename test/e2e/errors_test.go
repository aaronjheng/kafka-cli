package e2e_test

import (
	"testing"
)

func TestTopicDescribe_NonExistentTopic(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "describe", "nonexistent-topic-e2e")
	if err == nil {
		t.Error("expected error when describing non-existent topic")
	}
}

func TestTopicDelete_NonExistentTopic(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "delete", "nonexistent-topic-e2e")
	if err == nil {
		t.Error("expected error when deleting non-existent topic")
	}
}

func TestTopicAlter_NonExistentTopic(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "alter", "nonexistent-topic-e2e", "--partitions", "5")
	if err == nil {
		t.Error("expected error when altering non-existent topic")
	}
}

func TestTopicAlter_ReducePartitions(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "5", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = cli.Run(t.Context(), "topic", "delete", topic)
	})

	_, err = cli.Run(t.Context(), "topic", "alter", topic, "--partitions", "2")
	if err == nil {
		t.Error("expected error when reducing partition count")
	}
}

func TestTopicGetOffsets_NonExistentTopic(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "topic", "get-offsets", "nonexistent-topic-e2e")
	if err == nil {
		t.Error("expected error when getting offsets for non-existent topic")
	}
}

func TestGroupDescribe_NonExistentGroup(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "group", "describe", "nonexistent-group-e2e")
	if err == nil {
		t.Error("expected error when describing non-existent consumer group")
	}
}

func TestGroupDelete_NonExistentGroup(t *testing.T) {
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

	_, err := cli.Run(t.Context(), "group", "delete", "nonexistent-group-e2e")
	if err == nil {
		t.Error("expected error when deleting non-existent consumer group")
	}
}
