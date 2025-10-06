package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "group",
	}

	cmd.AddCommand(groupListCmd)

	return cmd
}

var groupListCmd = &cobra.Command{
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

		consumerGroups := []string{}
		for k := range groups {
			consumerGroups = append(consumerGroups, k)
		}

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
