package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(configCatCmd())

	return cmd
}

func configCatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "cat",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPathname := cfg.Filepath()
			fmt.Fprintf(os.Stdout, "# %s\n", cfgPathname)

			f, err := os.Open(cfgPathname)
			if err != nil {
				return fmt.Errorf("os.Open error: %w", err)
			}
			defer f.Close()

			if _, err := io.Copy(os.Stdout, f); err != nil {
				return fmt.Errorf("io.Copy error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
