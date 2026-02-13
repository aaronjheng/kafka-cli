package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"

	kafkago "github.com/segmentio/kafka-go"
)

var (
	errEmptyBrokers               = errors.New("brokers is empty")
	errTopicHasNoPartitions       = errors.New("topic has no partitions")
	errPartitionIDOutOfInt32Range = errors.New("partition id out of int32 range")
)

func ListTopicPartitions(ctx context.Context, brokers []string, dialer *kafkago.Dialer, topic string) ([]int32, error) {
	if len(brokers) == 0 {
		return nil, errEmptyBrokers
	}

	conn, err := dialer.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return nil, fmt.Errorf("dialer.DialContext error: %w", err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			slog.Error("conn.Close failed", slog.Any("error", err))
		}
	}()

	partitionMetas, err := conn.ReadPartitions(topic)
	if err != nil {
		return nil, fmt.Errorf("conn.ReadPartitions error: %w", err)
	}

	partitions := make([]int32, 0, len(partitionMetas))
	for _, partition := range partitionMetas {
		if partition.Topic != topic {
			continue
		}

		if partition.ID < math.MinInt32 || partition.ID > math.MaxInt32 {
			return nil, fmt.Errorf("%w: %d", errPartitionIDOutOfInt32Range, partition.ID)
		}

		partitions = append(partitions, int32(partition.ID))
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("%w: %s", errTopicHasNoPartitions, topic)
	}

	slices.Sort(partitions)

	return partitions, nil
}
