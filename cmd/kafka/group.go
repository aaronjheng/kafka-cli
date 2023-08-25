package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "group",
	Run: func(cmd *cobra.Command, args []string) {
	},
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
				slog.Error("clusterAdmin.Close failed", err)
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
		table.SetHeader([]string{"Consumer Group", "State", "Protocol Type", "Protocol", "Members"})

		for _, detail := range details {
			table.Append([]string{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, fmt.Sprintf("%d", len(detail.Members))})
		}

		table.Render()

		return nil
	},
}
