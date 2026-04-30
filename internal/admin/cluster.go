package admin

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
)

func (a *Admin) DescribeCluster() error {
	brokers, controllerID, err := a.clusterAdmin.DescribeCluster()
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeCluster error: %w", err)
	}

	topics, err := a.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListTopics error: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Brokers: %d\n", len(brokers))
	fmt.Fprintf(os.Stdout, "Topics: %d\n", len(topics))
	fmt.Fprintln(os.Stdout)

	return renderBrokerTable(brokers, controllerID)
}

func renderBrokerTable(brokers []*sarama.Broker, controllerID int32) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"ID", "Address", "Rack", "Type"})

	slices.SortStableFunc(brokers, func(a, b *sarama.Broker) int {
		return cmp.Compare(a.ID(), b.ID())
	})

	for _, broker := range brokers {
		rack := broker.Rack()
		if rack == "" {
			rack = "-"
		}

		brokerType := "broker"
		if broker.ID() == controllerID {
			brokerType = "controller"
		}

		err := table.Append([]any{broker.ID(), broker.Addr(), rack, brokerType})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	err := table.Render()
	if err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	return nil
}
