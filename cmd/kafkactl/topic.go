package main

import (
	"fmt"
	"log"
	"os"

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
