package admin

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
)

var errClusterIDNotAvailable = errors.New("cluster ID not available")

func (a *Admin) clusterID() (string, error) {
	broker, err := a.client.Controller()
	if err != nil {
		return "", fmt.Errorf("client.Controller error: %w", err)
	}

	request := sarama.NewMetadataRequest(a.client.Config().Version, nil)

	response, err := broker.GetMetadata(request)
	if err != nil {
		return "", fmt.Errorf("broker.GetMetadata error: %w", err)
	}

	if response.ClusterID == nil {
		return "", errClusterIDNotAvailable
	}

	return *response.ClusterID, nil
}

func (a *Admin) DescribeCluster() error {
	clusterID, err := a.clusterID()
	if err != nil {
		return fmt.Errorf("clusterID error: %w", err)
	}

	brokers, controllerID, err := a.clusterAdmin.DescribeCluster()
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeCluster error: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Cluster ID: %s\n\n", clusterID)
	fmt.Fprintln(os.Stdout, "Brokers")

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
