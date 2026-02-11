package admin

import (
	"cmp"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
)

func (a *Admin) ListTopics() error {
	topicDetails, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	topics := slices.SortedStableFunc(maps.Keys(topicDetails), func(a, b string) int {
		return cmp.Compare(a, b)
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Topic", "Number of Partitions", "Replication Factor"})

	for _, topic := range topics {
		topicDetail := topicDetails[topic]

		err := table.Append([]string{
			topic,
			fmt.Sprintf("%d", topicDetail.NumPartitions),
			fmt.Sprintf("%d", topicDetail.ReplicationFactor),
		})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

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

func (a *Admin) DeleteTopics(topics ...string) error {
	for _, topic := range topics {
		slog.Info("Delete topic", slog.String("topic", topic))

		if err := a.clusterAdmin.DeleteTopic(topic); err != nil {
			return fmt.Errorf("clusterAdmin.DeleteTopic error: %w", err)
		}

		slog.Info("Topic deleted", slog.String("topic", topic))
	}
	return nil
}
