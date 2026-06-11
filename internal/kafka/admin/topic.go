package admin

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
)

const offsetLookupConcurrency = 16

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

	tbl := newTable()
	tbl.Headers("Topic", "Number of Partitions", "Replication Factor")

	for _, topic := range topics {
		topicDetail := topicDetails[topic]

		tbl.Row(
			topic,
			strconv.FormatInt(int64(topicDetail.NumPartitions), 10),
			strconv.FormatInt(int64(topicDetail.ReplicationFactor), 10),
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

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
	tbl := newTable()
	tbl.Headers("Partition", "Leader", "Replicas", "ISR")

	slices.SortStableFunc(partitions, func(a, b *sarama.PartitionMetadata) int {
		return cmp.Compare(a.ID, b.ID)
	})

	for _, partition := range partitions {
		tbl.Row(
			strconv.FormatInt(int64(partition.ID), 10),
			strconv.FormatInt(int64(partition.Leader), 10),
			fmt.Sprint(partition.Replicas),
			fmt.Sprint(partition.Isr),
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

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

func (a *Admin) GetOffsets(topic string) error {
	partitions, err := a.topicPartitions(topic)
	if err != nil {
		return err
	}

	tbl := newTable()
	tbl.Headers("Partition", "Oldest Offset", "Newest Offset")

	for _, partition := range partitions {
		oldest, err := a.client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			return fmt.Errorf("client.GetOffset error: %w", err)
		}

		newest, err := a.client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("client.GetOffset error: %w", err)
		}

		tbl.Row(
			strconv.FormatInt(int64(partition), 10),
			strconv.FormatInt(oldest, 10),
			strconv.FormatInt(newest, 10),
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

	return nil
}

func (a *Admin) topicPartitions(topic string) ([]int32, error) {
	topicDetails, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	if _, ok := topicDetails[topic]; !ok {
		return nil, fmt.Errorf("%w: %s", errTopicNotFound, topic)
	}

	partitions, err := a.client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("client.Partitions error: %w", err)
	}

	slices.Sort(partitions)

	return partitions, nil
}

func (a *Admin) topicEndOffsets(topic string, partitions []int32) (map[int32]int64, error) {
	endOffsets := make(map[int32]int64, len(partitions))
	if len(partitions) == 0 {
		return endOffsets, nil
	}

	jobs := make(chan int32)
	results := make(chan offsetLookupResult, len(partitions))
	workerCount := min(len(partitions), offsetLookupConcurrency)

	var waitGroup sync.WaitGroup

	for range workerCount {
		waitGroup.Go(func() {
			for partition := range jobs {
				endOffset, err := a.client.GetOffset(topic, partition, sarama.OffsetNewest)
				results <- offsetLookupResult{
					partition: partition,
					offset:    endOffset,
					err:       err,
				}
			}
		})
	}

	go func() {
		for _, partition := range partitions {
			jobs <- partition
		}

		close(jobs)
		waitGroup.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("client.GetOffset error: %w", result.err)
		}

		endOffsets[result.partition] = result.offset
	}

	return endOffsets, nil
}

type offsetLookupResult struct {
	partition int32
	offset    int64
	err       error
}

func summarizeOffsetsLag(
	partitions []int32,
	topicOffsets map[int32]*sarama.OffsetFetchResponseBlock,
	endOffsets map[int32]int64,
) (int64, int64, int64, bool) {
	var (
		currentOffset      int64
		logEndOffset       int64
		lag                int64
		hasCommittedOffset bool
	)

	for _, partition := range partitions {
		block, ok := topicOffsets[partition]
		if !ok || block == nil || block.Offset < 0 {
			continue
		}

		endOffset := endOffsets[partition]
		partitionLag := max(endOffset-block.Offset, 0)

		currentOffset += block.Offset
		logEndOffset += endOffset
		lag += partitionLag
		hasCommittedOffset = true
	}

	return currentOffset, logEndOffset, lag, hasCommittedOffset
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
