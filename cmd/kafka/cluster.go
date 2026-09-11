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

	cmd.AddCommand(
		clusterDescribeCmd(meta),
		clusterConfigCmd(meta),
	)

	return cmd
}

func clusterConfigCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Kafka broker configs",
	}

	cmd.AddCommand(clusterConfigDescribeCmd(meta))

	return cmd
}

func clusterConfigDescribeCmd(meta *Meta) *cobra.Command {
	var (
		brokerID int32
		all      bool
	)

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show broker configs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				if brokerID < 0 {
					return a.DescribeAllBrokerConfigs(all)
				}

				return a.DescribeBrokerConfig(brokerID)
			})
		},
	}

	cmd.Flags().Int32Var(&brokerID, "broker", -1, "Broker ID, negative means all brokers")
	cmd.Flags().BoolVar(&all, "all", false, "Show all configs instead of only overrides")
	cmd.MarkFlagsMutuallyExclusive("all", "broker")

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
