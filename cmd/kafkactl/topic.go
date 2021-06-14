package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var topicCmd = &cobra.Command{
	Use: "topic",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var topicListCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		brokers, clusterCfg, err := cfg.Cluster(profile)
		if err != nil {
			log.Fatal(err)
		}

		cluster, err := sarama.NewClusterAdmin(brokers, clusterCfg)
		if err != nil {
			log.Fatal(err)
		}
		defer cluster.Close()

		topics, err := cluster.ListTopics()
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
