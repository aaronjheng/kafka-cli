package main

import (
	"cmp"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "group",
	}

	cmd.AddCommand(groupListCmd())
	cmd.AddCommand(groupDeleteCmd())

	return cmd
}

func groupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterAdmin, err := newClusterAdmin()
			if err != nil {
				return fmt.Errorf("newClusterAdmin error: %w", err)
			}

			defer func() {
				if err := clusterAdmin.Close(); err != nil {
					slog.Error("clusterAdmin.Close failed", slog.Any("error", err))
				}
			}()

			groups, err := clusterAdmin.ListConsumerGroups()
			if err != nil {
				return fmt.Errorf("clusterAdmin.ListConsumerGroups error: %w", err)
			}

			consumerGroups := slices.SortedStableFunc(maps.Keys(groups), func(a, b string) int {
				return cmp.Compare(a, b)
			})

			details, err := clusterAdmin.DescribeConsumerGroups(consumerGroups)
			if err != nil {
				return fmt.Errorf("clusterAdmin.DescribeConsumerGroups error: %w", err)
			}

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
		},
	}

	return cmd
}

func groupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterAdmin, err := newClusterAdmin()
			if err != nil {
				return fmt.Errorf("newClusterAdmin error: %w", err)
			}

			defer func() {
				if err := clusterAdmin.Close(); err != nil {
					slog.Error("clusterAdmin.Close failed", slog.Any("error", err))
				}
			}()

			group := args[0]
			slog.Info("Delete consumer group", slog.String("group", group))

			if err := clusterAdmin.DeleteConsumerGroup(group); err != nil {
				return fmt.Errorf("clusterAdmin.DeleteConsumerGroup error: %w", err)
			}

			slog.Info("Consumer group deleted", slog.String("group", group))

			return nil
		},
	}

	return cmd
}
