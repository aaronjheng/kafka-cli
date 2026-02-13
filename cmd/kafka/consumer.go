package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

const consumerMessageChannelBufferSize = 10

type consumerMessage struct {
	partition int
	value     string
}

func consumerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "consumer",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	cmd.AddCommand(consumerConsoleCmd())

	return cmd
}

func consumerConsoleCmd() *cobra.Command {
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

			partition, err := cmd.Flags().GetInt32("partition")
			if err != nil {
				return fmt.Errorf("get partition flag error: %w", err)
			}

			dialer, err := kafka.NewDialer(clusterCfg)
			if err != nil {
				return fmt.Errorf("kafka.NewDialer error: %w", err)
			}

			// Partition flag not specified.
			var partitions []int32

			if partition == -1 {
				partitions, err = kafka.ListTopicPartitions(cmd.Context(), clusterCfg.Brokers, dialer, topic)
				if err != nil {
					return fmt.Errorf("kafka.ListTopicPartitions error: %w", err)
				}
			} else {
				partitions = []int32{partition}
			}

			partitionWidth := 1

			for _, partition := range partitions {
				width := len(strconv.Itoa(int(partition)))
				if width > partitionWidth {
					partitionWidth = width
				}
			}

			var wg sync.WaitGroup

			msgCh := make(chan consumerMessage, consumerMessageChannelBufferSize)

			for _, partition := range partitions {
				wg.Go(func() {
					reader, err := kafka.NewPartitionReader(clusterCfg.Brokers, dialer, topic, partition)
					if err != nil {
						slog.Error("kafka.NewPartitionReader failed", slog.Any("error", err))

						return
					}

					defer func() {
						err := reader.Close()
						if err != nil {
							slog.Error("reader.Close failed", slog.Any("error", err))
						}
					}()

					err = reader.SetOffset(kafkago.LastOffset)
					if err != nil {
						slog.Error("reader.SetOffset failed", slog.Any("error", err))

						return
					}

					for {
						msg, err := reader.ReadMessage(cmd.Context())
						if err != nil {
							if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
								return
							}

							slog.Error("reader.ReadMessage failed", slog.Any("error", err))

							return
						}

						msgCh <- consumerMessage{
							partition: msg.Partition,
							value:     string(msg.Value),
						}
					}
				})
			}

			go func() {
				wg.Wait()
				close(msgCh)
			}()

			for msg := range msgCh {
				_, err := fmt.Fprintf(os.Stdout, "[%0*d] %s\n", partitionWidth, msg.partition, msg.value)
				if err != nil {
					slog.Error("fmt.Fprintf failed", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("topic", "t", "", "The topic to consume from")
	cmd.Flags().Int32P("partition", "p", -1, "The partition to consume from.")

	return cmd
}
