package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// TODO: kafka-topics --config

func topicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "topic",
		Run: func(cmd *cobra.Command, args []string) {
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

			topics, err := clusterAdmin.ListTopics()
			if err != nil {
				return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Topic", "Number of Partitions", "Replication Factor"})

			for k, v := range topics {
				table.Append([]string{k, fmt.Sprintf("%d", v.NumPartitions), fmt.Sprintf("%d", v.ReplicationFactor)})
			}

			table.Render()

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
			topic := args[0]

			clusterAdmin, err := newClusterAdmin()
			if err != nil {
				return fmt.Errorf("newClusterAdmin error: %w", err)
			}

			defer func() {
				if err := clusterAdmin.Close(); err != nil {
					slog.Error("clusterAdmin.Close failed", slog.Any("error", err))
				}
			}()

			numPartitions, err := cmd.Flags().GetInt32("partitions")
			if err != nil {
				return fmt.Errorf("get partitions flag error: %w", err)
			}

			replicationFactor, err := cmd.Flags().GetInt16("replication-factor")
			if err != nil {
				return fmt.Errorf("get replication-factor flag error: %w", err)
			}

			err = clusterAdmin.CreateTopic(topic, &sarama.TopicDetail{
				NumPartitions:     numPartitions,
				ReplicationFactor: replicationFactor,
			}, false)
			if err != nil {
				return fmt.Errorf("clusterAdmin.CreateTopic error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int32("partitions", 1, "The number of partitions for the topic")
	cmd.Flags().Int16("replication-factor", 1, "The replication factor for each partition in the topic being created.")

	return cmd
}

func topicDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "delete",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := args[0]

			clusterAdmin, err := newClusterAdmin()
			if err != nil {
				return fmt.Errorf("newClusterAdmin error: %w", err)
			}

			defer func() {
				if err := clusterAdmin.Close(); err != nil {
					slog.Error("clusterAdmin.Close failed", slog.Any("error", err))
				}
			}()

			if err := clusterAdmin.DeleteTopic(topic); err != nil {
				return fmt.Errorf("clusterAdmin.DeleteTopic error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
