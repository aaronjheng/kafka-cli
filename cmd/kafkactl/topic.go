package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// TODO: kafka-topics --create
// TODO: kafka-topics --delete
// TODO: kafka-topics --config

var topicCmd = &cobra.Command{
	Use:   "topic",
	Short: "topic",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var topicListCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		cluserAdmin, err := newClusterAdmin()
		if err != nil {
			log.Fatal(err)
		}

		topics, err := cluserAdmin.ListTopics()
		if err != nil {
			log.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Topic", "Number of Partitions", "Replication Factor"})

		for k, v := range topics {
			table.Append([]string{k, fmt.Sprintf("%d", v.NumPartitions), fmt.Sprintf("%d", v.ReplicationFactor)})
		}

		table.Render()

	},
}

var topicCreateCmd = &cobra.Command{
	Use:  "create",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		topic := args[0]

		cluserAdmin, err := newClusterAdmin()
		if err != nil {
			log.Fatal(err)
		}

		numPartitions, err := cmd.Flags().GetInt32("partitions")
		if err != nil {
			log.Fatal(err)
		}

		replicationFactor, err := cmd.Flags().GetInt16("replication-factor")
		if err != nil {
			log.Fatal(err)
		}

		err = cluserAdmin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		}, false)
		if err != nil {
			log.Fatal(err)
		}
	},
}
