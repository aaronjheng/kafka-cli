package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	var (
		cfgFilepath string
		cluster     string
	)

	meta := NewMeta(&cfgFilepath, &cluster)

	cmd := &cobra.Command{
		Use:          "kafka",
		Short:        "Command line tool for Apache Kafka",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if _, ok := cmd.Annotations["skipConfigLoad"]; ok {
				return nil
			}

			_, err := meta.Config()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			return nil
		},
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Cluster name to operate.")
	cmd.PersistentFlags().StringVarP(&cfgFilepath, "config", "f", "", "Config file path.")

	err := cmd.RegisterFlagCompletionFunc("cluster", clusterCompletionFunc(meta))
	if err != nil {
		panic(fmt.Sprintf("RegisterFlagCompletionFunc error: %v", err))
	}

	cmd.AddCommand(configCmd(meta))
	cmd.AddCommand(clusterCmd(meta))
	cmd.AddCommand(topicCmd(meta))
	cmd.AddCommand(groupCmd(meta))
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(completionCmd())

	return cmd
}

func main() {
	// Bootstrap logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
