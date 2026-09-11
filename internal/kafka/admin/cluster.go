package admin

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/IBM/sarama"
)

var errClusterIDNotAvailable = errors.New("cluster ID not available")

var (
	errBrokerConfigQueryFailed          = errors.New("broker config query failed")
	errUnexpectedDescribeConfigsResults = errors.New("unexpected number of DescribeConfigs results")
)

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

func (a *Admin) DescribeBrokerConfig(brokerID int32) error {
	entries, err := a.brokerConfigEntries(brokerID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Broker: %d\n", brokerID)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Configs")

	renderConfigEntries(entries)

	return nil
}

func (a *Admin) DescribeAllBrokerConfigs(showAll bool) error {
	ids, err := a.brokerIDs()
	if err != nil {
		return err
	}

	resources := make([]*sarama.ConfigResource, 0, len(ids))
	for _, id := range ids {
		resources = append(resources, &sarama.ConfigResource{
			Type: sarama.BrokerResource,
			Name: strconv.FormatInt(int64(id), 10),
		})
	}

	results, err := a.clusterAdmin.DescribeConfigs(resources, sarama.DescribeConfigsOptions{})
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeConfigs error: %w", err)
	}

	for resultIdx, result := range results {
		err := renderBrokerConfigs(result, showAll)
		if err != nil {
			return err
		}

		if resultIdx < len(results)-1 {
			fmt.Fprintln(os.Stdout)
		}
	}

	return nil
}

func renderBrokerConfigs(result *sarama.ConfigResourceResult, showAll bool) error {
	if result.ErrorCode != sarama.ErrNoError {
		return fmt.Errorf("broker %s: %s: %w", result.Name, result.ErrorMsg, errBrokerConfigQueryFailed)
	}

	entries := result.Configs
	if !showAll {
		entries = configOverrides(result.Configs)
	}

	slices.SortStableFunc(entries, func(x, y sarama.ConfigEntry) int {
		return cmp.Compare(x.Name, y.Name)
	})

	fmt.Fprintf(os.Stdout, "Broker: %s\n", result.Name)
	fmt.Fprintln(os.Stdout)

	sectionTitle := "Config Overrides"
	if showAll {
		sectionTitle = "Configs"
	}

	fmt.Fprintln(os.Stdout, sectionTitle)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stdout, "No config overrides")
	} else {
		renderConfigEntries(entries)
	}

	return nil
}

func configOverrides(entries []sarama.ConfigEntry) []sarama.ConfigEntry {
	overrides := make([]sarama.ConfigEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Source != sarama.SourceDefault {
			overrides = append(overrides, entry)
		}
	}

	return overrides
}

func (a *Admin) brokerIDs() ([]int32, error) {
	brokers, _, err := a.clusterAdmin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.DescribeCluster error: %w", err)
	}

	ids := make([]int32, 0, len(brokers))
	for _, broker := range brokers {
		ids = append(ids, broker.ID())
	}

	slices.Sort(ids)

	return ids, nil
}

func (a *Admin) brokerConfigEntries(brokerID int32) ([]sarama.ConfigEntry, error) {
	results, err := a.clusterAdmin.DescribeConfigs([]*sarama.ConfigResource{{
		Type: sarama.BrokerResource,
		Name: strconv.FormatInt(int64(brokerID), 10),
	}}, sarama.DescribeConfigsOptions{})
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.DescribeConfigs error: %w", err)
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("%w: %d", errUnexpectedDescribeConfigsResults, len(results))
	}

	result := results[0]
	if result.ErrorCode != sarama.ErrNoError {
		return nil, fmt.Errorf("broker %d: %s: %w", brokerID, result.ErrorMsg, errBrokerConfigQueryFailed)
	}

	slices.SortStableFunc(result.Configs, func(x, y sarama.ConfigEntry) int {
		return cmp.Compare(x.Name, y.Name)
	})

	return result.Configs, nil
}

func renderConfigEntries(entries []sarama.ConfigEntry) {
	tbl := newTable()
	tbl.Headers("Name", "Value", "Source", "Read Only", "Sensitive")

	for _, entry := range entries {
		value := entry.Value
		if value == "" {
			value = "-"
		}

		if entry.Sensitive {
			value = "*****"
		}

		tbl.Row(
			entry.Name,
			value,
			configSourceName(entry.Source),
			strconv.FormatBool(entry.ReadOnly),
			strconv.FormatBool(entry.Sensitive),
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())
}

func configSourceName(source sarama.ConfigSource) string {
	switch source {
	case sarama.SourceUnknown:
		return "unknown"
	case sarama.SourceTopic:
		return "dynamic-topic"
	case sarama.SourceDynamicBroker:
		return "dynamic-broker"
	case sarama.SourceDynamicDefaultBroker:
		return "dynamic-default-broker"
	case sarama.SourceStaticBroker:
		return "static-broker"
	case sarama.SourceDefault:
		return "default"
	default:
		return "unknown"
	}
}

func renderBrokerTable(brokers []*sarama.Broker, controllerID int32) error {
	tbl := newTable()
	tbl.Headers("ID", "Address", "Rack", "Type")

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

		tbl.Row(strconv.FormatInt(int64(broker.ID()), 10), broker.Addr(), rack, brokerType)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

	return nil
}
