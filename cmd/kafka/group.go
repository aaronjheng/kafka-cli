package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage consumer groups",
	}

	cmd.AddCommand(groupListCmd())
	cmd.AddCommand(groupDescribeCmd())
	cmd.AddCommand(groupDeleteCmd())

	return cmd
}

func groupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumer groups",
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
					// Ignore error
				}
			}()

			err = admin.ListConsumerGroups()
			if err != nil {
				return fmt.Errorf("admin.ListConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func groupDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe GROUP",
		Short:             "Describe a consumer group",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			err = admin.DescribeConsumerGroup(args[0])
			if err != nil {
				return fmt.Errorf("admin.DescribeConsumerGroup error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func groupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete consumer groups",
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.DeleteConsumerGroups(args...)
			if err != nil {
				return fmt.Errorf("admin.DeleteConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
