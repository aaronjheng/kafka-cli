package main

import (
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "group",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "list",
	Run: func(cmd *cobra.Command, args []string) {
		cluserAdmin, err := newClusterAdmin()
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := cluserAdmin.Close(); err != nil {
				log.Println(err)
			}
		}()

		groups, err := cluserAdmin.ListConsumerGroups()
		if err != nil {
			log.Fatal(err)
		}

		consumerGroups := []string{}
		for k := range groups {
			consumerGroups = append(consumerGroups, k)
		}

		details, err := cluserAdmin.DescribeConsumerGroups(consumerGroups)
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
