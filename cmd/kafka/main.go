package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/config"
	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

var (
	cfg     *config.Config
	cluster string
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Cluster name to operate.")
	cmd.PersistentFlags().StringP("config", "f", "", "Config file path.")

	cmd.AddCommand(configCmd())
	cmd.AddCommand(topicCmd())
	cmd.AddCommand(groupCmd())
	cmd.AddCommand(producerCmd())
	cmd.AddCommand(consumerCmd())
	cmd.AddCommand(versionCmd)
	cmd.AddCommand(completionCmd)

	return cmd
}

func main() {
	// Bootstrap logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCluster() (*kafka.Kafka, error) {
	cfg, err := cfg.Cluster(cluster)
	if err != nil {
		return nil, fmt.Errorf("cfg.Cluster error: %w", err)
	}

	return kafka.New(cfg)
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
