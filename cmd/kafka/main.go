package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/admin"
	"github.com/aaronjheng/kafka-cli/internal/config"
	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

//nolint:gochecknoglobals // Cobra command wiring keeps shared CLI state here.
var (
	cfg     *config.Config
	cluster string
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kafka",
		Short:        "Command line tool for Apache Kafka",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
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
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(completionCmd())

	return cmd
}

func main() {
	// Bootstrap logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func newCluster() (*kafka.Kafka, error) {
	clusterCfg, err := clusterConfig()
	if err != nil {
		return nil, fmt.Errorf("clusterConfig error: %w", err)
	}

	cluster, err := kafka.New(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka.New error: %w", err)
	}

	return cluster, nil
}

func clusterConfig() (*kafka.Config, error) {
	clusterCfg, err := cfg.Cluster(cluster)
	if err != nil {
		return nil, fmt.Errorf("cfg.Cluster error: %w", err)
	}

	return clusterCfg, nil
}

func provideAdmin() (*admin.Admin, func(context.Context) error, error) {
	cluster, err := newCluster()
	if err != nil {
		return nil, nil, fmt.Errorf("newCluster error: %w", err)
	}

	clusterAdmin, err := sarama.NewClusterAdminFromClient(cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("newClusterAdmin error: %w", err)
	}

	return admin.NewAdmin(clusterAdmin), func(_ context.Context) error {
		return clusterAdmin.Close()
	}, nil
}
