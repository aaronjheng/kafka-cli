package main

import (
	"bufio"
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

const (
	defaultTopicPartitions      int32 = 3
	consumerMessageChannelBufSz int   = 10
)

type consumerMessage struct {
	partition int
	value     string
}

func topicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "topic",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	cmd.AddCommand(topicListCmd())
	cmd.AddCommand(topicCreateCmd())
	cmd.AddCommand(topicDeleteCmd())
	cmd.AddCommand(topicDescribeCmd())
	cmd.AddCommand(topicConsumeCmd())
	cmd.AddCommand(topicProduceCmd())

	return cmd
}

func topicListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.ListTopics()
			if err != nil {
				return fmt.Errorf("admin.ListTopics error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			topic := args[0]

			numPartitions, err := cmd.Flags().GetInt32("partitions")
			if err != nil {
				return fmt.Errorf("get partitions flag error: %w", err)
			}

			replicationFactor, err := cmd.Flags().GetInt16("replication-factor")
			if err != nil {
				return fmt.Errorf("get replication-factor flag error: %w", err)
			}

			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.CreateTopic(topic, numPartitions, replicationFactor)
			if err != nil {
				return fmt.Errorf("admin.CreateTopic error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int32("partitions", defaultTopicPartitions, "The number of partitions for the topic")
	cmd.Flags().Int16("replication-factor", 1, "The replication factor for each partition in the topic being created.")

	return cmd
}

func topicDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
					// Ignore error
				}
			}()

			err = admin.DeleteTopics(args...)
			if err != nil {
				return fmt.Errorf("admin.DeleteTopics error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe TOPIC",
		Short: "Show details of a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			admin, closer, err := provideAdmin()
			if err != nil {
				return fmt.Errorf("provideAdmin error: %w", err)
			}

			defer func() {
				err := closer(ctx)
				if err != nil {
					slog.Error("closer error", slog.Any("error", err))
				}
			}()

			err = admin.DescribeTopic(args[0])
			if err != nil {
				return fmt.Errorf("admin.DescribeTopic error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func topicConsumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "consume TOPIC",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterCfg, err := clusterConfig()
			if err != nil {
				return fmt.Errorf("clusterConfig error: %w", err)
			}

			topic := args[0]

			partition, err := cmd.Flags().GetInt32("partition")
			if err != nil {
				return fmt.Errorf("get partition flag error: %w", err)
			}

			var partitions []int32
			if partition == -1 {
				partitions, err = kafka.ListTopicPartitions(cmd.Context(), clusterCfg, topic)
				if err != nil {
					return fmt.Errorf("kafka.ListTopicPartitions error: %w", err)
				}
			} else {
				partitions = []int32{partition}
			}

			partitionWidth := calculatePartitionWidth(partitions)
			msgCh := startPartitionReaders(cmd.Context(), clusterCfg, topic, partitions)

			printMessages(msgCh, partitionWidth)

			return nil
		},
	}

	cmd.Flags().Int32P("partition", "p", -1, "The partition to consume from.")

	return cmd
}

func topicProduceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "produce TOPIC",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			clusterCfg, err := clusterConfig()
			if err != nil {
				return fmt.Errorf("clusterConfig error: %w", err)
			}

			topic := args[0]

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

	return cmd
}

func calculatePartitionWidth(partitions []int32) int {
	partitionWidth := 1

	for _, partition := range partitions {
		width := len(strconv.Itoa(int(partition)))
		if width > partitionWidth {
			partitionWidth = width
		}
	}

	return partitionWidth
}

func startPartitionReaders(
	ctx context.Context,
	clusterCfg *kafka.Config,
	topic string,
	partitions []int32,
) <-chan consumerMessage {
	var waitGroup sync.WaitGroup

	msgCh := make(chan consumerMessage, consumerMessageChannelBufSz)

	for _, partition := range partitions {
		waitGroup.Go(func() {
			readPartitionMessages(ctx, clusterCfg, topic, partition, msgCh)
		})
	}

	go func() {
		waitGroup.Wait()
		close(msgCh)
	}()

	return msgCh
}

func readPartitionMessages(
	ctx context.Context,
	clusterCfg *kafka.Config,
	topic string,
	partition int32,
	msgCh chan<- consumerMessage,
) {
	reader, err := kafka.NewPartitionReader(clusterCfg, topic, partition)
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
		msg, err := reader.ReadMessage(ctx)
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
}

func printMessages(msgCh <-chan consumerMessage, partitionWidth int) {
	for msg := range msgCh {
		_, err := fmt.Fprintf(os.Stdout, "[%0*d] %s\n", partitionWidth, msg.partition, msg.value)
		if err != nil {
			slog.Error("fmt.Fprintf failed", slog.Any("error", err))
		}
	}
}
