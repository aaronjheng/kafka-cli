package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/config"
)

var (
	cfg     *config.Config
	cluster string
)

var rootCmd = &cobra.Command{
	Use:          "kafka",
	Short:        "Command line tool for Apache Kafka",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfgFilepath, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("config flag error: %w", err)
		}

		cfg, err = config.LoadConfig(cfgFilepath)
		if err != nil {
			return fmt.Errorf("config flag error: %w", err)
		}

		return nil
	},
}

func main() {
	// Bootstrap logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	rootCmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Cluster name to operate.")
	rootCmd.PersistentFlags().StringP("config", "f", "", "Config file path.")

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCatCmd)

	rootCmd.AddCommand(topicCmd)
	topicCmd.AddCommand(topicListCmd)
	topicCmd.AddCommand(topicCreateCmd)
	topicCreateCmd.Flags().Int32("partitions", 1, "The number of partitions for the topic")
	topicCreateCmd.Flags().Int16("replication-factor", 1, "The replication factor for each partition in the topic being created.")
	topicCmd.AddCommand(topicDeleteCmd)

	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(groupListCmd)

	rootCmd.AddCommand(producerCmd)
	producerCmd.AddCommand(producerConsoleCmd)
	producerConsoleCmd.Flags().StringP("topic", "t", "", "The topic to produce messages to.")

	rootCmd.AddCommand(consumerCmd)
	consumerCmd.AddCommand(consumerConsoleCmd)
	consumerConsoleCmd.Flags().StringP("topic", "t", "", "The topic to consume from")
	consumerConsoleCmd.Flags().Int32P("partition", "p", -1, "The partition to consume from.")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCluster() (sarama.Client, error) {
	brokers, clusterCfg, err := cfg.Cluster(cluster)
	if err != nil {
		return nil, fmt.Errorf("cfg.Cluster error: %w", err)
	}

	return sarama.NewClient(brokers, clusterCfg)
}

func newClusterAdmin() (sarama.ClusterAdmin, error) {
	cluster, err := newCluster()
	if err != nil {
		return nil, fmt.Errorf("newCluster error: %w", err)
	}

	return sarama.NewClusterAdminFromClient(cluster)
}

func newSyncProducer() (sarama.SyncProducer, error) {
	cluster, err := newCluster()
	if err != nil {
		return nil, fmt.Errorf("newCluster error: %w", err)
	}

	cluster.Config().Producer.Return.Successes = true

	return sarama.NewSyncProducerFromClient(cluster)
}

func newConsumer() (sarama.Consumer, error) {
	cluster, err := newCluster()
	if err != nil {
		return nil, fmt.Errorf("newCluster error: %w", err)
	}

	return sarama.NewConsumerFromClient(cluster)
}
