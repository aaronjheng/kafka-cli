package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "group",
	}

	cmd.AddCommand(groupListCmd())
	cmd.AddCommand(groupDeleteCmd())

	return cmd
}

func groupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				if err := closer(ctx); err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			if err := admin.ListConsumerGroups(); err != nil {
				return fmt.Errorf("admin.ListConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func groupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				if err := closer(ctx); err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			if err := admin.DeleteConsumerGroups(args...); err != nil {
				return fmt.Errorf("admin.DeleteConsumerGroups error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
