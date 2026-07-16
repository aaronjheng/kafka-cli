package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func clusterCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kafka cluster",
	}

	cmd.AddCommand(clusterDescribeCmd(meta))

	return cmd
}

func clusterDescribeCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of the cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				err := a.DescribeCluster()
				if err != nil {
					return fmt.Errorf("admin.DescribeCluster error: %w", err)
				}

				return nil
			})
		},
	}

	return cmd
}
