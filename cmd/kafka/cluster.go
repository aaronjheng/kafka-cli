package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func clusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kafka cluster",
	}

	cmd.AddCommand(clusterDescribeCmd())

	return cmd
}

func clusterDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of the cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			admin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
				}
			}()

			err = admin.DescribeCluster()
			if err != nil {
				return fmt.Errorf("admin.DescribeCluster error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
