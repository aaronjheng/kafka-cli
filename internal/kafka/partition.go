package kafka

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

var errTopicHasNoPartitions = errors.New("topic has no partitions")

func ListTopicPartitions(ctx context.Context, cfg *Config, topic string) ([]int32, error) {
	_ = ctx

	client, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("newSaramaClient error: %w", err)
	}
	defer client.Close()

	partitions, err := client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("client.Partitions error: %w", err)
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("%w: %s", errTopicHasNoPartitions, topic)
	}

	slices.Sort(partitions)

	return partitions, nil
}
