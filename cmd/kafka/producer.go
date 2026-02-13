package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

func producerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "producer",
		Short: "producer",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	cmd.AddCommand(producerConsoleCmd())

	return cmd
}

func producerConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "console",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clusterCfg, err := clusterConfig()
			if err != nil {
				return fmt.Errorf("clusterConfig error: %w", err)
			}

			topic, err := cmd.Flags().GetString("topic")
			if err != nil {
				return fmt.Errorf("get topic flag error: %w", err)
			}

			writer, err := kafka.NewWriter(clusterCfg, topic)
			if err != nil {
				return fmt.Errorf("kafka.NewWriter error: %w", err)
			}

			defer func() {
				err := writer.Close()
				if err != nil {
					slog.Error("writer.Close failed", slog.Any("error", err))
				}
			}()

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()

				if line == "" {
					continue
				}

				err := writer.WriteMessages(context.Background(), kafkago.Message{
					Value: []byte(line),
				})
				if err != nil {
					slog.Error("writer.WriteMessages failed", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("topic", "t", "", "The topic to produce messages to.")

	return cmd
}
