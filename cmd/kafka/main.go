package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/config"
)

//nolint:gochecknoglobals // Cobra command wiring keeps shared CLI state here.
var (
	cfg     *config.Config
	cluster string
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kafka",
		Short:        "Command line tool for Apache Kafka",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfgFilepath, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("config flag error: %w", err)
			}

			cfg, err = config.LoadConfig(cfgFilepath)
			if err != nil {
				return fmt.Errorf("config flag error: %w", err)
			}

			return nil
		},
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Cluster name to operate.")
	cmd.PersistentFlags().StringP("config", "f", "", "Config file path.")

	err := cmd.RegisterFlagCompletionFunc("cluster", clusterCompletionFunc)
	if err != nil {
		panic(fmt.Sprintf("RegisterFlagCompletionFunc error: %v", err))
	}

	cmd.AddCommand(configCmd())
	cmd.AddCommand(clusterCmd())
	cmd.AddCommand(topicCmd())
	cmd.AddCommand(groupCmd())
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
