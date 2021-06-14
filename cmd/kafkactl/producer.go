package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var producerCmd = &cobra.Command{
	Use:   "producer",
	Short: "producer",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var producerConsoleCmd = &cobra.Command{
	Use: "console",
	RunE: func(cmd *cobra.Command, args []string) error {
		producer, err := newSyncProducer()
		if err != nil {
			return fmt.Errorf("newSyncProducer error: %w", err)
		}

		defer func() {
			if err := producer.Close(); err != nil {
				logger.Error("producer.Close failed", zap.Error(err))
			}
		}()

		topic, err := cmd.Flags().GetString("topic")
		if err != nil {
			return fmt.Errorf("get topic flag error: %w", err)
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(scanner.Text()),
			}
			_, _, err := producer.SendMessage(msg)
			if err != nil {
				logger.Error("producer.SendMessage failed", zap.Error(err))
			}
		}

		return nil
	},
}
