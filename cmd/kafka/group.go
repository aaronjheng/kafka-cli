package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage consumer groups",
	}

	cmd.AddCommand(groupListCmd())
	cmd.AddCommand(groupDescribeCmd())
	cmd.AddCommand(groupOffsetsCmd())
	cmd.AddCommand(groupLagCmd())
	cmd.AddCommand(groupDeleteCmd())

	return cmd
}

func groupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumer groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			admin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.ListConsumerGroups()
			if err != nil {
				return fmt.Errorf("admin.ListConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func groupDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe GROUP",
		Short:             "Describe a consumer group",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
				}
			}()

			err = admin.DescribeConsumerGroup(args[0])
			if err != nil {
				return fmt.Errorf("admin.DescribeConsumerGroup error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func groupOffsetsCmd() *cobra.Command {
	return newGroupTopicCmd(
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

func groupLagCmd() *cobra.Command {
	return newGroupTopicCmd(
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
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			kafkaAdmin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
				}
			}()

			err = run(kafkaAdmin, args[0], topic)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&topic, "topic", "", topicFlagUsage)

	err := cmd.RegisterFlagCompletionFunc("topic", topicCompletionFunc)
	if err != nil {
		panic(fmt.Sprintf("RegisterFlagCompletionFunc error: %v", err))
	}

	return cmd
}

func groupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete consumer groups",
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := admin.NewFromConfig(cfg, cluster)
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.DeleteConsumerGroups(args...)
			if err != nil {
				return fmt.Errorf("admin.DeleteConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
