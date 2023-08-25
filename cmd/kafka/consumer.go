package main

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
)

var consumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "consumer",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var consumerConsoleCmd = &cobra.Command{
	Use: "console",
	RunE: func(cmd *cobra.Command, args []string) error {
		consumer, err := newConsumer()
		if err != nil {
			return fmt.Errorf("newConsumer error: %w", err)
		}

		defer func() {
			if err := consumer.Close(); err != nil {
				slog.Error("consumer.Close failed", err)
			}
		}()

		topic, err := cmd.Flags().GetString("topic")
		if err != nil {
			return fmt.Errorf("get topic flag error: %w", err)
		}

		partition, err := cmd.Flags().GetInt32("partition")
		if err != nil {
			return fmt.Errorf("get partition flag error: %w", err)
		}

		// Partition flag not specified
		var partitions []int32
		if partition == -1 {
			var err error
			partitions, err = consumer.Partitions(topic)
			if err != nil {
				return fmt.Errorf("consumer.Partitions error: %w", err)
			}
		} else {
			partitions = []int32{partition}
		}

		var wg sync.WaitGroup

		msgCh := make(chan string, 10)

		for _, partition := range partitions {
			wg.Add(1)
			go func(partition int32) {
				defer wg.Done()

				partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
				if err != nil {
					slog.Error("consumer.ConsumePartition failed", err)
					return
				}

				for msg := range partitionConsumer.Messages() {
					msgCh <- string(msg.Value)
				}
			}(partition)
		}

		go func() {
			for msg := range msgCh {
				fmt.Println(msg)
			}
		}()

		wg.Wait()

		return nil
	},
}
