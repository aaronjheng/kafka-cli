package admin

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
)

func (a *Admin) ListTopicNames() ([]string, error) {
	topicDetails, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	return slices.SortedStableFunc(maps.Keys(topicDetails), cmp.Compare), nil
}

func (a *Admin) ListTopics() error {
	topicDetails, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	topics := slices.SortedStableFunc(maps.Keys(topicDetails), cmp.Compare)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"Topic", "Number of Partitions", "Replication Factor"})

	for _, topic := range topics {
		topicDetail := topicDetails[topic]

		err := table.Append([]any{
			topic,
			topicDetail.NumPartitions,
			topicDetail.ReplicationFactor,
		})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	err = table.Render()
	if err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Total topics: %d\n", len(topics))

	return nil
}

func (a *Admin) CreateTopic(topic string, numPartitions int32, replicationFactor int16) error {
	err := a.clusterAdmin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}, false)
	if err != nil {
		return fmt.Errorf("clusterAdmin.CreateTopic error: %w", err)
	}

	return nil
}

const unknownValue = "unknown"

var (
	errTopicNotFound         = errors.New("topic not found")
	errTopicMetadataNotFound = errors.New("topic metadata not found")
)

func (a *Admin) DescribeTopic(topic string) error {
	topicDetails, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	detail, ok := topicDetails[topic]
	if !ok {
		return fmt.Errorf("%w: %s", errTopicNotFound, topic)
	}

	metadata, err := a.clusterAdmin.DescribeTopics([]string{topic})
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeTopics error: %w", err)
	}

	if len(metadata) == 0 {
		return fmt.Errorf("%w: %s", errTopicMetadataNotFound, topic)
	}

	topicMeta := metadata[0]

	cleanupPolicy := a.getCleanupPolicy(topic)
	size := a.getTopicSize(topic)

	fmt.Fprintf(os.Stdout, "Topic: %s\n", topic)
	fmt.Fprintf(os.Stdout, "Partitions: %d\n", detail.NumPartitions)
	fmt.Fprintf(os.Stdout, "Replication Factor: %d\n", detail.ReplicationFactor)
	fmt.Fprintf(os.Stdout, "Cleanup Policy: %s\n", cleanupPolicy)
	fmt.Fprintf(os.Stdout, "Size: %s\n", size)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Partitions")

	return a.renderPartitionTable(topicMeta.Partitions)
}

func (a *Admin) renderPartitionTable(partitions []*sarama.PartitionMetadata) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"Partition", "Leader", "Replicas", "ISR"})

	for _, partition := range partitions {
		err := table.Append([]any{
			partition.ID,
			partition.Leader,
			partition.Replicas,
			partition.Isr,
		})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	err := table.Render()
	if err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	return nil
}

func (a *Admin) getCleanupPolicy(topic string) string {
	configEntries, err := a.clusterAdmin.DescribeConfig(sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topic,
	})
	if err != nil {
		slog.Debug("DescribeConfig failed", slog.String("topic", topic), slog.Any("error", err))

		return unknownValue
	}

	for _, entry := range configEntries {
		if entry.Name == "cleanup.policy" {
			return entry.Value
		}
	}

	return unknownValue
}

func (a *Admin) getTopicSize(topic string) string {
	brokers, _, err := a.clusterAdmin.DescribeCluster()
	if err != nil {
		slog.Debug("DescribeCluster failed", slog.String("topic", topic), slog.Any("error", err))

		return unknownValue
	}

	brokerIDs := make([]int32, len(brokers))
	for i, b := range brokers {
		brokerIDs[i] = b.ID()
	}

	logDirs, err := a.clusterAdmin.DescribeLogDirs(brokerIDs)
	if err != nil {
		slog.Debug("DescribeLogDirs failed", slog.String("topic", topic), slog.Any("error", err))

		return unknownValue
	}

	var totalSize int64

	for _, dirs := range logDirs {
		for _, dir := range dirs {
			for _, t := range dir.Topics {
				if t.Topic == topic {
					for _, p := range t.Partitions {
						totalSize += p.Size
					}
				}
			}
		}
	}

	return formatSize(totalSize)
}

func formatSize(bytes int64) string {
	const (
		kiloByte = 1024
		megaByte = kiloByte * 1024
		gigaByte = megaByte * 1024
	)

	switch {
	case bytes >= gigaByte:
		return fmt.Sprintf("%.2f GB", float64(bytes)/gigaByte)
	case bytes >= megaByte:
		return fmt.Sprintf("%.2f MB", float64(bytes)/megaByte)
	case bytes >= kiloByte:
		return fmt.Sprintf("%.2f KB", float64(bytes)/kiloByte)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (a *Admin) AlterTopicPartitions(topic string, numPartitions int32) error {
	err := a.clusterAdmin.CreatePartitions(topic, numPartitions, nil, false)
	if err != nil {
		return fmt.Errorf("clusterAdmin.CreatePartitions error: %w", err)
	}

	return nil
}

func (a *Admin) DeleteTopics(topics ...string) error {
	for _, topic := range topics {
		slog.Info("Delete topic", slog.String("topic", topic))

		err := a.clusterAdmin.DeleteTopic(topic)
		if err != nil {
			return fmt.Errorf("clusterAdmin.DeleteTopic error: %w", err)
		}

		slog.Info("Topic deleted", slog.String("topic", topic))
	}

	return nil
}
