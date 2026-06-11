package admin

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strconv"

	"github.com/IBM/sarama"
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

	tbl := newTable()
	tbl.Headers("Consumer Group", "State", "Protocol Type", "Protocol", "Members")

	for _, detail := range details {
		tbl.Row(detail.GroupId, detail.State, detail.ProtocolType, detail.Protocol, strconv.Itoa(len(detail.Members)))
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

	return nil
}

func (a *Admin) DescribeConsumerGroup(group string) error {
	desc, err := a.describeConsumerGroup(group)
	if err != nil {
		return err
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

	return nil
}

func (a *Admin) describeConsumerGroup(group string) (*sarama.GroupDescription, error) {
	details, err := a.clusterAdmin.DescribeConsumerGroups([]string{group})
	if err != nil {
		return nil, fmt.Errorf("clusterAdmin.DescribeConsumerGroups error: %w", err)
	}

	if len(details) == 0 || details[0].State == "Dead" {
		return nil, fmt.Errorf("%w: %s", errConsumerGroupNotFound, group)
	}

	return details[0], nil
}

func (a *Admin) ListConsumerGroupOffsets(group string, topic string) error {
	desc, offsetResp, partitionsByTopic, err := a.consumerGroupOffsets(group, topic)
	if err != nil {
		return err
	}

	partitions, endOffsets, err := a.buildGroupOffsetLookups(offsetResp.Blocks, partitionsByTopic)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Consumer Group: %s\n", desc.GroupId)
	fmt.Fprintln(os.Stdout)

	assignments := a.buildPartitionAssignments(desc.Members)
	renderGroupOffsetsTable(offsetResp.Blocks, assignments, partitions, endOffsets)

	return nil
}

func (a *Admin) ListConsumerGroupLag(group string, topic string) error {
	desc, offsetResp, partitionsByTopic, err := a.consumerGroupOffsets(group, topic)
	if err != nil {
		return err
	}

	partitions, endOffsets, err := a.buildGroupOffsetLookups(offsetResp.Blocks, partitionsByTopic)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Consumer Group: %s\n", desc.GroupId)
	fmt.Fprintln(os.Stdout)

	renderGroupLagSummary(offsetResp.Blocks, partitions, endOffsets)

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Partitions")

	assignments := a.buildPartitionAssignments(desc.Members)
	renderGroupOffsetsTable(offsetResp.Blocks, assignments, partitions, endOffsets)

	return nil
}

func (a *Admin) consumerGroupOffsets(
	group string,
	topic string,
) (*sarama.GroupDescription, *sarama.OffsetFetchResponse, topicPartitionIDs, error) {
	desc, err := a.describeConsumerGroup(group)
	if err != nil {
		return nil, nil, nil, err
	}

	partitionsByTopic := make(topicPartitionIDs)

	var requestedPartitions map[string][]int32

	if topic != "" {
		partitions, err := a.topicPartitions(topic)
		if err != nil {
			return nil, nil, nil, err
		}

		partitionsByTopic[topic] = partitions
		requestedPartitions = partitionsByTopic
	}

	offsetResp, err := a.clusterAdmin.ListConsumerGroupOffsets(group, requestedPartitions)
	if err != nil {
		if errors.Is(err, sarama.ErrGroupIDNotFound) {
			return nil, nil, nil, fmt.Errorf("%w: %s", errConsumerGroupNotFound, group)
		}

		return nil, nil, nil, fmt.Errorf("clusterAdmin.ListConsumerGroupOffsets error: %w", err)
	}

	return desc, offsetResp, partitionsByTopic, nil
}

func (a *Admin) buildGroupOffsetLookups(
	blocks topicPartitionOffsets,
	requestedPartitions topicPartitionIDs,
) (topicPartitionIDs, topicPartitionEndOffsets, error) {
	partitionsByTopic := make(topicPartitionIDs, len(blocks))
	endOffsetsByTopic := make(topicPartitionEndOffsets, len(blocks))
	topics := slices.SortedStableFunc(maps.Keys(blocks), cmp.Compare)

	for _, topic := range topics {
		partitions := requestedPartitions[topic]
		if len(partitions) == 0 {
			partitions = slices.SortedStableFunc(maps.Keys(blocks[topic]), cmp.Compare)
		}

		endOffsets, err := a.topicEndOffsets(topic, partitions)
		if err != nil {
			return nil, nil, err
		}

		partitionsByTopic[topic] = partitions
		endOffsetsByTopic[topic] = endOffsets
	}

	return partitionsByTopic, endOffsetsByTopic, nil
}

func (a *Admin) renderGroupMembersTable(members map[string]*sarama.GroupMemberDescription) error {
	tbl := newTable()
	tbl.Headers("Member ID", "Client ID", "Client Host")

	for _, member := range members {
		tbl.Row(member.MemberId, member.ClientId, member.ClientHost)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())

	return nil
}

type (
	topicPartitionOffsets    = map[string]map[int32]*sarama.OffsetFetchResponseBlock
	topicPartitionOwners     = map[string]map[int32]partitionOwner
	topicPartitionIDs        = map[string][]int32
	topicPartitionEndOffsets = map[string]map[int32]int64
	memberDescriptions       = map[string]*sarama.GroupMemberDescription
	partitionOffsetResponse  = map[int32]*sarama.OffsetFetchResponseBlock
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

func renderGroupLagSummary(
	blocks topicPartitionOffsets,
	partitionsByTopic topicPartitionIDs,
	endOffsetsByTopic topicPartitionEndOffsets,
) {
	tbl := newTable()
	tbl.Headers("Topic", "Current Offset", "Log End Offset", "Lag")

	topics := slices.SortedStableFunc(maps.Keys(blocks), cmp.Compare)

	for _, topic := range topics {
		currentOffset, logEndOffset, lag, ok := summarizeOffsetsLag(
			partitionsByTopic[topic],
			blocks[topic],
			endOffsetsByTopic[topic],
		)
		if !ok {
			continue
		}

		tbl.Row(
			topic,
			strconv.FormatInt(currentOffset, 10),
			strconv.FormatInt(logEndOffset, 10),
			strconv.FormatInt(lag, 10),
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())
}

func renderGroupOffsetsTable(
	blocks topicPartitionOffsets,
	assignments topicPartitionOwners,
	partitionsByTopic topicPartitionIDs,
	endOffsetsByTopic topicPartitionEndOffsets,
) {
	topics := slices.SortedStableFunc(maps.Keys(blocks), cmp.Compare)

	for i, topic := range topics {
		if i > 0 {
			fmt.Fprintln(os.Stdout)
		}

		fmt.Fprintf(os.Stdout, "Topic: %s\n", topic)

		renderTopicOffsetsTable(
			blocks[topic],
			partitionsByTopic[topic],
			assignments[topic],
			endOffsetsByTopic[topic],
		)
	}
}

func renderTopicOffsetsTable(
	partitions partitionOffsetResponse,
	partitionIDs []int32,
	owners map[int32]partitionOwner,
	endOffsets map[int32]int64,
) {
	tbl := newTable()
	tbl.Headers("Partition", "Current Offset", "Log End Offset", "Lag", "Consumer ID", "Host")

	for _, partitionID := range partitionIDs {
		block := partitions[partitionID]
		if block == nil || block.Offset < 0 {
			continue
		}

		endOffset := endOffsets[partitionID]
		lag := max(endOffset-block.Offset, 0)

		var consumerID, host string

		if owner, ok := owners[partitionID]; ok {
			consumerID = owner.memberID
			host = owner.clientHost
		}

		tbl.Row(
			strconv.FormatInt(int64(partitionID), 10),
			strconv.FormatInt(block.Offset, 10),
			strconv.FormatInt(endOffset, 10),
			strconv.FormatInt(lag, 10),
			consumerID,
			host,
		)
	}

	fmt.Fprintln(os.Stdout, tbl.Render())
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
