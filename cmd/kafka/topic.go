package main

import (
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// TODO: kafka-topics --config

var topicCmd = &cobra.Command{
	Use:   "topic",
	Short: "topic",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var topicListCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterAdmin, err := newClusterAdmin()
		if err != nil {
			return fmt.Errorf("newClusterAdmin error: %w", err)
		}

		defer func() {
			if err := clusterAdmin.Close(); err != nil {
				logger.Error("clusterAdmin.Close failed", zap.Error(err))
			}
		}()

		topics, err := clusterAdmin.ListTopics()
		if err != nil {
			return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Topic", "Number of Partitions", "Replication Factor"})

		for k, v := range topics {
			table.Append([]string{k, fmt.Sprintf("%d", v.NumPartitions), fmt.Sprintf("%d", v.ReplicationFactor)})
		}

		table.Render()

		return nil
	},
}

var topicCreateCmd = &cobra.Command{
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
				logger.Error("clusterAdmin.Close failed", zap.Error(err))
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

var topicDeleteCmd = &cobra.Command{
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
				logger.Error("clusterAdmin.Close failed", zap.Error(err))
			}
		}()

		if err := clusterAdmin.DeleteTopic(topic); err != nil {
			return fmt.Errorf("clusterAdmin.DeleteTopic error: %w", err)
		}

		return nil
	},
}