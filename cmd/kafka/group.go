package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
		cluserAdmin, err := newClusterAdmin()
		if err != nil {
			return fmt.Errorf("newClusterAdmin error: %w", err)
		}

		defer func() {
			if err := cluserAdmin.Close(); err != nil {
				logger.Error("cluserAdmin.Close failed", zap.Error(err))
			}
		}()

		groups, err := cluserAdmin.ListConsumerGroups()
		if err != nil {
			return fmt.Errorf("cluserAdmin.ListConsumerGroups error: %w", err)
		}

		consumerGroups := []string{}
		for k := range groups {
			consumerGroups = append(consumerGroups, k)
		}

		details, err := cluserAdmin.DescribeConsumerGroups(consumerGroups)
		if err != nil {
			return fmt.Errorf("cluserAdmin.DescribeConsumerGroups error: %w", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Consumer Group", "State", "Protocal Type", "Protocal", "Members"})

		for _, detail := range details {
			table.Append([]string{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, fmt.Sprintf("%d", len(detail.Members))})
		}

		table.Render()

		return nil
	},
}
