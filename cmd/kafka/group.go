package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func groupCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage consumer groups",
	}

	cmd.AddCommand(groupListCmd(meta))
	cmd.AddCommand(groupDescribeCmd(meta))
	cmd.AddCommand(groupOffsetsCmd(meta))
	cmd.AddCommand(groupLagCmd(meta))
	cmd.AddCommand(groupDeleteCmd(meta))

	return cmd
}

func groupListCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumer groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				err := a.ListConsumerGroups()
				if err != nil {
					return fmt.Errorf("admin.ListConsumerGroups error: %w", err)
				}

				return nil
			})
		},
	}

	return cmd
}

func groupDescribeCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe GROUP",
		Short:             "Describe a consumer group",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: consumerGroupCompletionFunc(meta),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				err := a.DescribeConsumerGroup(args[0])
				if err != nil {
					return fmt.Errorf("admin.DescribeConsumerGroup error: %w", err)
				}

				return nil
			})
		},
	}

	return cmd
}

func groupOffsetsCmd(meta *Meta) *cobra.Command {
	return newGroupTopicCmd(
		meta,
		"offsets GROUP",
		"Show committed offsets for a consumer group",
		"Only show offsets for the specified topic",
		func(admin *admin.Admin, group string, topic string) error {
			err := admin.ListConsumerGroupOffsets(group, topic)
			if err != nil {
				return fmt.Errorf("admin.ListConsumerGroupOffsets error: %w", err)
			}

			return nil
		},
	)
}

func groupLagCmd(meta *Meta) *cobra.Command {
	return newGroupTopicCmd(
		meta,
		"lag GROUP",
		"Show lag for a consumer group",
		"Only show lag for the specified topic",
		func(admin *admin.Admin, group string, topic string) error {
			err := admin.ListConsumerGroupLag(group, topic)
			if err != nil {
				return fmt.Errorf("admin.ListConsumerGroupLag error: %w", err)
			}

			return nil
		},
	)
}

type groupTopicRunFunc func(admin *admin.Admin, group string, topic string) error

func newGroupTopicCmd(
	meta *Meta,
	use string,
	short string,
	topicFlagUsage string,
	run groupTopicRunFunc,
) *cobra.Command {
	var topic string

	cmd := &cobra.Command{
		Use:               use,
		Short:             short,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: consumerGroupCompletionFunc(meta),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				return run(a, args[0], topic)
			})
		},
	}

	cmd.Flags().StringVar(&topic, "topic", "", topicFlagUsage)

	err := cmd.RegisterFlagCompletionFunc("topic", topicCompletionFunc(meta))
	if err != nil {
		panic(fmt.Sprintf("RegisterFlagCompletionFunc error: %v", err))
	}

	return cmd
}

func groupDeleteCmd(meta *Meta) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete consumer groups",
		ValidArgsFunction: consumerGroupCompletionFunc(meta),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withAdmin(cmd.Context(), meta, func(a *admin.Admin) error {
				err := a.DeleteConsumerGroups(args...)
				if err != nil {
					return fmt.Errorf("admin.DeleteConsumerGroups error: %w", err)
				}

				return nil
			})
		},
	}

	return cmd
}
