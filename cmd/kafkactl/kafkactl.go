package main

import (
	"log"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafkactl/pkg/config"
)

var cfg *config.Config
var cluster string

var rootCmd = &cobra.Command{
	Use:   "kafkactl",
	Short: "Command line tool for Apache Kafka",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfgFilepath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatal(err)
		}

		cfg, err = config.LoadConfig(cfgFilepath)
		if err != nil {
			log.Fatal(err)
		}
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
		log.Fatal(err)
	}
}

func newCluster() (sarama.Client, error) {
	brokers, clusterCfg, err := cfg.Cluster(cluster)
	if err != nil {
		log.Fatal(err)
	}

	return sarama.NewClient(brokers, clusterCfg)
}

func newClusterAdmin() (sarama.ClusterAdmin, error) {
	cluster, err := newCluster()
	if err != nil {
		log.Fatal(err)
	}

	return sarama.NewClusterAdminFromClient(cluster)
}

func newSyncProducer() (sarama.SyncProducer, error) {
	cluster, err := newCluster()
	if err != nil {
		log.Fatal(err)
	}

	cluster.Config().Producer.Return.Successes = true

	return sarama.NewSyncProducerFromClient(cluster)
}

func newConsumer() (sarama.Consumer, error) {
	cluster, err := newCluster()
	if err != nil {
		log.Fatal(err)
	}

	return sarama.NewConsumerFromClient(cluster)
}
