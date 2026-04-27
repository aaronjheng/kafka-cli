package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	errExactlyOneResetStrategy = errors.New("exactly one of --to-earliest, --to-latest, or --to-offset must be specified")
	errTopicRequired           = errors.New("at least one --topic must be specified")
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage consumer groups",
	}

	cmd.AddCommand(groupListCmd())
	cmd.AddCommand(groupDeleteCmd())
	cmd.AddCommand(groupResetOffsetsCmd())

	return cmd
}

func groupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumer groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
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

func groupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete consumer groups",
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
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

func validateResetOffsetFlags(cmd *cobra.Command) (bool, bool, int64, error) {
	toEarliest, err := cmd.Flags().GetBool("to-earliest")
	if err != nil {
		return false, false, 0, fmt.Errorf("get to-earliest flag error: %w", err)
	}

	toLatest, err := cmd.Flags().GetBool("to-latest")
	if err != nil {
		return false, false, 0, fmt.Errorf("get to-latest flag error: %w", err)
	}

	toOffset, err := cmd.Flags().GetInt64("to-offset")
	if err != nil {
		return false, false, 0, fmt.Errorf("get to-offset flag error: %w", err)
	}

	strategyCount := 0
	if toEarliest {
		strategyCount++
	}

	if toLatest {
		strategyCount++
	}

	if cmd.Flags().Changed("to-offset") {
		strategyCount++
	}

	if strategyCount != 1 {
		return false, false, 0, errExactlyOneResetStrategy
	}

	return toEarliest, toLatest, toOffset, nil
}

func groupResetOffsetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "reset-offsets GROUP",
		Short:             "Reset consumer group offsets",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: consumerGroupCompletionFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			group := args[0]

			topics, err := cmd.Flags().GetStringSlice("topic")
			if err != nil {
				return fmt.Errorf("get topic flag error: %w", err)
			}

			if len(topics) == 0 {
				return errTopicRequired
			}

			toEarliest, toLatest, toOffset, err := validateResetOffsetFlags(cmd)
			if err != nil {
				return err
			}

			admin, closer, err := provideAdmin()
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

			err = admin.ResetConsumerGroupOffsets(group, topics, toEarliest, toLatest, toOffset)
			if err != nil {
				return fmt.Errorf("admin.ResetConsumerGroupOffsets error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringSlice("topic", nil, "Topics to reset offsets for")
	cmd.Flags().Bool("to-earliest", false, "Reset offsets to earliest")
	cmd.Flags().Bool("to-latest", false, "Reset offsets to latest")
	cmd.Flags().Int64("to-offset", -1, "Reset offsets to a specific value")

	_ = cmd.RegisterFlagCompletionFunc("topic", topicCompletionFunc)

	return cmd
}
