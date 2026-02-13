package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

const defaultTopicPartitions int32 = 3

func topicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "topic",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	cmd.AddCommand(topicListCmd())
	cmd.AddCommand(topicCreateCmd())
	cmd.AddCommand(topicDeleteCmd())

	return cmd
}

func topicListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
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

			err = admin.ListTopics()
			if err != nil {
				return fmt.Errorf("admin.ListTopics error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			topic := args[0]

			numPartitions, err := cmd.Flags().GetInt32("partitions")
			if err != nil {
				return fmt.Errorf("get partitions flag error: %w", err)
			}

			replicationFactor, err := cmd.Flags().GetInt16("replication-factor")
			if err != nil {
				return fmt.Errorf("get replication-factor flag error: %w", err)
			}

			admin, closer, err := provideAdmin()
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

			err = admin.CreateTopic(topic, numPartitions, replicationFactor)
			if err != nil {
				return fmt.Errorf("admin.CreateTopic error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int32("partitions", defaultTopicPartitions, "The number of partitions for the topic")
	cmd.Flags().Int16("replication-factor", 1, "The replication factor for each partition in the topic being created.")

	return cmd
}

func topicDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
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

			err = admin.DeleteTopics(args...)
			if err != nil {
				return fmt.Errorf("admin.DeleteTopics error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
