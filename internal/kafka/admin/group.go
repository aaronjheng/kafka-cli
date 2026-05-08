package admin

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
)

func (a *Admin) ListConsumerGroupIDs() ([]string, error) {
	groups, err := a.clusterAdmin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.ListConsumerGroups error: %w", err)
	}

	return slices.SortedStableFunc(maps.Keys(groups), cmp.Compare), nil
}

func (a *Admin) ListConsumerGroups() error {
	groups, err := a.clusterAdmin.ListConsumerGroups()
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListConsumerGroups error: %w", err)
	}

	details, err := a.clusterAdmin.DescribeConsumerGroups(slices.Collect(maps.Keys(groups)))
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeConsumerGroups error: %w", err)
	}

	slices.SortStableFunc(details, func(a, b *sarama.GroupDescription) int {
		return cmp.Compare(a.GroupId, b.GroupId)
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Consumer Group", "State", "Protocol Type", "Protocol", "Members"})

	for _, detail := range details {
		err := table.Append([]any{detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, len(detail.Members)})
		if err != nil {
			return fmt.Errorf("table.Append error: %w", err)
		}
	}

	err = table.Render()
	if err != nil {
		return fmt.Errorf("table.Render error: %w", err)
	}

	return nil
}

func (a *Admin) DescribeConsumerGroup(group string) error {
	details, err := a.clusterAdmin.DescribeConsumerGroups([]string{group})
	if err != nil {
		return fmt.Errorf("clusterAdmin.DescribeConsumerGroups error: %w", err)
	}

	if len(details) == 0 {
		return fmt.Errorf("%w: %s", errConsumerGroupNotFound, group)
	}

	desc := details[0]

	if desc.State == "Dead" {
		return fmt.Errorf("%w: %s", errConsumerGroupNotFound, group)
	}

	offsetResp, err := a.clusterAdmin.ListConsumerGroupOffsets(group, nil)
	if err != nil {
		return fmt.Errorf("clusterAdmin.ListConsumerGroupOffsets error: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Consumer Group: %s\n", desc.GroupId)
	fmt.Fprintf(os.Stdout, "State: %s\n", desc.State)
	fmt.Fprintf(os.Stdout, "Protocol Type: %s\n", desc.ProtocolType)
	fmt.Fprintf(os.Stdout, "Protocol: %s\n", desc.Protocol)
	fmt.Fprintf(os.Stdout, "Members: %d\n", len(desc.Members))
	fmt.Fprintln(os.Stdout)

	if len(desc.Members) > 0 {
		fmt.Fprintln(os.Stdout, "Members")

		err = a.renderGroupMembersTable(desc.Members)
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout)
	}

	if len(offsetResp.Blocks) > 0 {
		fmt.Fprintln(os.Stdout, "Offsets")

		assignments := a.buildPartitionAssignments(desc.Members)

		err = a.renderGroupOffsetsTable(offsetResp.Blocks, assignments)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Admin) renderGroupMembersTable(members map[string]*sarama.GroupMemberDescription) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"Member ID", "Client ID", "Client Host"})

	for _, member := range members {
		err := table.Append([]any{member.MemberId, member.ClientId, member.ClientHost})
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

type (
	topicPartitionOffsets   = map[string]map[int32]*sarama.OffsetFetchResponseBlock
	topicPartitionOwners    = map[string]map[int32]partitionOwner
	memberDescriptions      = map[string]*sarama.GroupMemberDescription
	partitionOffsetResponse = map[int32]*sarama.OffsetFetchResponseBlock
)

type partitionOwner struct {
	memberID   string
	clientHost string
}

func (a *Admin) buildPartitionAssignments(members memberDescriptions) topicPartitionOwners {
	assignments := make(map[string]map[int32]partitionOwner)

	for _, member := range members {
		assignment, err := member.GetMemberAssignment()
		if err != nil {
			continue
		}

		for topic, partitionIDs := range assignment.Topics {
			if assignments[topic] == nil {
				assignments[topic] = make(map[int32]partitionOwner)
			}

			for _, partitionID := range partitionIDs {
				assignments[topic][partitionID] = partitionOwner{
					memberID:   member.MemberId,
					clientHost: member.ClientHost,
				}
			}
		}
	}

	return assignments
}

func (a *Admin) renderGroupOffsetsTable(blocks topicPartitionOffsets, assignments topicPartitionOwners) error {
	topics := slices.SortedStableFunc(maps.Keys(blocks), cmp.Compare)

	for i, topic := range topics {
		if i > 0 {
			fmt.Fprintln(os.Stdout)
		}

		fmt.Fprintf(os.Stdout, "Topic: %s\n", topic)

		partitions := blocks[topic]

		err := a.renderTopicOffsetsTable(topic, partitions, assignments[topic])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Admin) renderTopicOffsetsTable(
	topic string,
	partitions partitionOffsetResponse,
	owners map[int32]partitionOwner,
) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"Partition", "Current Offset", "Log End Offset", "Lag", "Consumer ID", "Host"})

	partitionIDs := slices.SortedStableFunc(maps.Keys(partitions), cmp.Compare)

	for _, partitionID := range partitionIDs {
		block := partitions[partitionID]

		endOffset, err := a.client.GetOffset(topic, partitionID, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("client.GetOffset error: %w", err)
		}

		lag := endOffset - block.Offset

		var consumerID, host string

		if owner, ok := owners[partitionID]; ok {
			consumerID = owner.memberID
			host = owner.clientHost
		}

		err = table.Append([]any{partitionID, block.Offset, endOffset, lag, consumerID, host})
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

var errConsumerGroupNotFound = errors.New("consumer group not found")

func (a *Admin) DeleteConsumerGroups(groups ...string) error {
	for _, group := range groups {
		slog.Info("Delete consumer group", slog.String("group", group))

		err := a.clusterAdmin.DeleteConsumerGroup(group)
		if err != nil {
			return fmt.Errorf("clusterAdmin.DeleteConsumerGroup error: %w", err)
		}

		slog.Info("Consumer group deleted", slog.String("group", group))
	}

	return nil
}
