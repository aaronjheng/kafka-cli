package main

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aaronjheng/kafkactl/pkg/config"
)

var cfg *config.Config
var cluster string
var logger *zap.Logger

var rootCmd = &cobra.Command{
	Use:   "kafkactl",
	Short: "Command line tool for Apache Kafka",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfgFilepath, err := cmd.Flags().GetString("config")
		if err != nil {
			logger.Fatal("config flag error", zap.Error(err))
		}

		cfg, err = config.LoadConfig(cfgFilepath)
		if err != nil {
			logger.Fatal("LoadConfig failed", zap.Error(err))
		}

		logger, _ = newLogger()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		_ = logger.Sync()
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Cluster name to operate.")
	rootCmd.PersistentFlags().StringP("config", "f", "", "Config file path.")

	rootCmd.AddCommand(completionCmd)

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

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("command failed", zap.Error(err))
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

func newLogger() (*zap.Logger, error) {
	cfg := &zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return cfg.Build()
}
