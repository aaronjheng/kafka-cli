package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

const defaultTopicPartitions int32 = 3

func topicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	cmd.AddCommand(topicListCmd())
	cmd.AddCommand(topicCreateCmd())
	cmd.AddCommand(topicAlterCmd())
	cmd.AddCommand(topicDeleteCmd())
	cmd.AddCommand(topicDescribeCmd())
	cmd.AddCommand(topicGetOffsetsCmd())
	cmd.AddCommand(topicConsumeCmd())
	cmd.AddCommand(topicProduceCmd())

	return cmd
}

func topicListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all topics",
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
		Use:   "create",
		Short: "Create a topic",
		Args:  cobra.ExactArgs(1),
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

func topicAlterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "alter TOPIC",
		Short:             "Alter a topic",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: topicCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			topic := args[0]

			numPartitions, err := cmd.Flags().GetInt32("partitions")
			if err != nil {
				return fmt.Errorf("get partitions flag error: %w", err)
			}

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

			err = admin.AlterTopicPartitions(topic, numPartitions)
			if err != nil {
				return fmt.Errorf("admin.AlterTopicPartitions error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int32("partitions", -1, "The total number of partitions for the topic")

	return cmd
}

func topicDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete topics",
		ValidArgsFunction: topicCompletionFunc,
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

			err = admin.DeleteTopics(args...)
			if err != nil {
				return fmt.Errorf("admin.DeleteTopics error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe TOPIC",
		Short:             "Show details of a topic",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: topicCompletionFunc,
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

			err = admin.DescribeTopic(args[0])
			if err != nil {
				return fmt.Errorf("admin.DescribeTopic error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicGetOffsetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get-offsets TOPIC",
		Short:             "Show oldest and newest offsets for each partition of a topic",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: topicCompletionFunc,
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

			err = admin.GetOffsets(args[0])
			if err != nil {
				return fmt.Errorf("admin.GetOffsets error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicConsumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "consume TOPIC",
		Short:             "Consume messages from a topic",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: topicCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterCfg, err := cfg.Cluster(cluster)
			if err != nil {
				return fmt.Errorf("clusterConfig error: %w", err)
			}

			topic := args[0]

			partition, err := cmd.Flags().GetInt32("partition")
			if err != nil {
				return fmt.Errorf("get partition flag error: %w", err)
			}

			var partitions []int32
			if partition == -1 {
				partitions, err = kafka.ListTopicPartitions(cmd.Context(), clusterCfg, topic)
				if err != nil {
					return fmt.Errorf("kafka.ListTopicPartitions error: %w", err)
				}
			} else {
				partitions = []int32{partition}
			}

			partitionWidth := kafka.CalculatePartitionWidth(partitions)
			msgCh := kafka.StartPartitionReaders(cmd.Context(), clusterCfg, topic, partitions)

			kafka.PrintMessages(msgCh, partitionWidth)

			return nil
		},
	}

	cmd.Flags().Int32P("partition", "p", -1, "The partition to consume from.")

	return cmd
}

func topicProduceCmd() *cobra.Command {
	var keySeparator string

	cmd := &cobra.Command{
		Use:               "produce TOPIC",
		Short:             "Produce messages to a topic",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: topicCompletionFunc,
		RunE: func(_ *cobra.Command, args []string) error {
			clusterCfg, err := cfg.Cluster(cluster)
			if err != nil {
				return fmt.Errorf("clusterConfig error: %w", err)
			}

			return kafka.RunTopicProduce(clusterCfg, args[0], keySeparator)
		},
	}

	cmd.Flags().StringVar(&keySeparator, "key-separator", "", "Separator to split key from value")

	return cmd
}
