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

func (a *Admin) ListConsumerGroupIDs() ([]string, error) {
	groups, err := a.clusterAdmin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.ListConsumerGroups error: %w", err)
	}

	return slices.SortedStableFunc(maps.Keys(groups), cmp.Compare), nil
}

func (a *Admin) ListConsumerGroups() error {
	groups, err := a.clusterAdmin.ListConsumerGroups()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListConsumerGroups error: %w", err)
	}

	details, err := a.clusterAdmin.DescribeConsumerGroups(slices.Collect(maps.Keys(groups)))
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeConsumerGroups error: %w", err)
	}

	slices.SortStableFunc(details, func(a, b *sarama.GroupDescription) int {
		return cmp.Compare(a.GroupId, b.GroupId)
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Consumer Group", "State", "Protocol Type", "Protocol", "Members"})

	for _, detail := range details {
		err := table.Append([]any{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, len(detail.Members)})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	err = table.Render()
	if err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	return nil
}

func (a *Admin) DeleteConsumerGroups(groups ...string) error {
	for _, group := range groups {
		slog.Info("Delete consumer group", slog.String("group", group))

		err := a.clusterAdmin.DeleteConsumerGroup(group)
		if err != nil {
			return fmt.Errorf("clusterAdmin.DeleteConsumerGroup error: %w", err)
		}

		slog.Info("Consumer group deleted", slog.String("group", group))
	}

	return nil
}

func (a *Admin) ResetConsumerGroupOffsets(
	group string,
	topics []string,
	toEarliest, toLatest bool,
	toOffset int64,
) error {
	offsetManager, err := sarama.NewOffsetManagerFromClient(group, a.client)
	if err != nil {
		return fmt.Errorf("sarama.NewOffsetManagerFromClient error: %w", err)
	}
	defer offsetManager.Close()

	for _, topic := range topics {
		partitions, err := a.client.Partitions(topic)
		if err != nil {
			return fmt.Errorf("client.Partitions error: %w", err)
		}

		for _, partition := range partitions {
			pom, err := offsetManager.ManagePartition(topic, partition)
			if err != nil {
				return fmt.Errorf("offsetManager.ManagePartition error: %w", err)
			}

			var targetOffset int64

			switch {
			case toEarliest:
				targetOffset, err = a.client.GetOffset(topic, partition, sarama.OffsetOldest)
			case toLatest:
				targetOffset, err = a.client.GetOffset(topic, partition, sarama.OffsetNewest)
			default:
				targetOffset = toOffset
			}

			if err != nil {
				return fmt.Errorf("client.GetOffset error: %w", err)
			}

			pom.ResetOffset(targetOffset, "")

			slog.Info("Reset offset",
				slog.String("group", group),
				slog.String("topic", topic),
				slog.Int("partition", int(partition)),
				slog.Int64("offset", targetOffset),
			)
		}
	}

	offsetManager.Commit()

	return nil
}
