package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
)

func producerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "producer",
		Short: "producer",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.AddCommand(producerConsoleCmd())

	return cmd
}

func producerConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "console",
		RunE: func(cmd *cobra.Command, args []string) error {
			producer, err := newSyncProducer()
			if err != nil {
				return fmt.Errorf("newSyncProducer error: %w", err)
			}

			defer func() {
				if err := producer.Close(); err != nil {
					slog.Error("producer.Close failed", slog.Any("error", err))
				}
			}()

			topic, err := cmd.Flags().GetString("topic")
			if err != nil {
				return fmt.Errorf("get topic flag error: %w", err)
			}

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()

				if line == "" {
					continue
				}

				msg := &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.StringEncoder(line),
				}
				_, _, err := producer.SendMessage(msg)
				if err != nil {
					slog.Error("producer.SendMessage failed", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("topic", "t", "", "The topic to produce messages to.")

	return cmd
}
