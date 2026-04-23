package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(configCatCmd())

	return cmd
}

func configCatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat",
		Short: "Print configuration file contents",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfgPathname := cfg.Filepath()
			fmt.Fprintf(os.Stdout, "# %s\n", cfgPathname)

			configFile, err := os.Open(cfgPathname)
			if err != nil {
				return fmt.Errorf("os.Open error: %w", err)
			}
			defer configFile.Close()

			_, err = io.Copy(os.Stdout, configFile)
			if err != nil {
				return fmt.Errorf("io.Copy error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
