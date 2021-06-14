package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use: "group",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var groupListCmd = &cobra.Command{
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

		groups, err := cluster.ListConsumerGroups()
		if err != nil {
			log.Fatal(err)
		}

		consumerGroups := []string{}
		for k := range groups {
			consumerGroups = append(consumerGroups, k)
		}

		details, err := cluster.DescribeConsumerGroups(consumerGroups)
		if err != nil {
			log.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Consumer Group", "State", "Protocal Type", "Protocal", "Members"})

		for _, detail := range details {
			table.Append([]string{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, fmt.Sprintf("%d", len(detail.Members))})
		}

		table.Render()

	},
}
