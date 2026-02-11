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

type Admin struct {
	clusterAdmin sarama.ClusterAdmin
}

func NewAdmin(clusterAdmin sarama.ClusterAdmin) *Admin {
	return &Admin{
		clusterAdmin: clusterAdmin,
	}
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
		err := table.Append([]string{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, fmt.Sprintf("%d", len(detail.Members))})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	return nil
}

func (a *Admin) DeleteConsumerGroups(groups ...string) error {
	for _, group := range groups {
		slog.Info("Delete consumer group", slog.String("group", group))

		if err := a.clusterAdmin.DeleteConsumerGroup(group); err != nil {
			return fmt.Errorf("clusterAdmin.DeleteConsumerGroup error: %w", err)
		}

		slog.Info("Consumer group deleted", slog.String("group", group))
	}

	return nil
}
