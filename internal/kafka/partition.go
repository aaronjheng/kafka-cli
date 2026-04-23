package kafka

import (
	"context"
	"errors"
	"fmt"
	"slices"

	kafkago "github.com/segmentio/kafka-go"
)

var errTopicHasNoPartitions = errors.New("topic has no partitions")

func ListTopicPartitions(ctx context.Context, cfg *Config, topic string) ([]int32, error) {
	transport, err := NewTransport(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewTransport error: %w", err)
	}

	client := &kafkago.Client{
		Addr:      kafkago.TCP(cfg.Brokers...),
		Transport: transport,
	}

	resp, err := client.Metadata(ctx, &kafkago.MetadataRequest{
		Topics: []string{topic},
	})
	if err != nil {
		return nil, fmt.Errorf("client.Metadata error: %w", err)
	}

	for _, topicMeta := range resp.Topics {
		if topicMeta.Name != topic {
			continue
		}

		partitions := make([]int32, 0, len(topicMeta.Partitions))
		for _, p := range topicMeta.Partitions {
			partitions = append(partitions, int32(p.ID)) //nolint:gosec // partition ID fits in int32
		}

		if len(partitions) == 0 {
			return nil, fmt.Errorf("%w: %s", errTopicHasNoPartitions, topic)
		}

		slices.Sort(partitions)

		return partitions, nil
	}

	return nil, fmt.Errorf("%w: %s", errTopicHasNoPartitions, topic)
}
