package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
)

const consumerMessageChannelBufSz = 10

type ConsumerMessage struct {
	Partition int
	Value     string
}

func CalculatePartitionWidth(partitions []int32) int {
	partitionWidth := 1

	for _, partition := range partitions {
		width := len(strconv.Itoa(int(partition)))
		if width > partitionWidth {
			partitionWidth = width
		}
	}

	return partitionWidth
}

func StartPartitionReaders(
	ctx context.Context,
	clusterCfg *Config,
	topic string,
	partitions []int32,
) <-chan ConsumerMessage {
	var waitGroup sync.WaitGroup

	msgCh := make(chan ConsumerMessage, consumerMessageChannelBufSz)

	for _, partition := range partitions {
		waitGroup.Go(func() {
			ReadPartitionMessages(ctx, clusterCfg, topic, partition, msgCh)
		})
	}

	go func() {
		waitGroup.Wait()
		close(msgCh)
	}()

	return msgCh
}

func ReadPartitionMessages(
	ctx context.Context,
	clusterCfg *Config,
	topic string,
	partition int32,
	msgCh chan<- ConsumerMessage,
) {
	reader, err := NewPartitionReader(clusterCfg, topic, partition, sarama.OffsetNewest)
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

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				return
			}

			slog.Error("reader.ReadMessage failed", slog.Any("error", err))

			return
		}

		msgCh <- ConsumerMessage{
			Partition: int(msg.Partition),
			Value:     string(msg.Value),
		}
	}
}

func PrintMessages(msgCh <-chan ConsumerMessage, partitionWidth int) {
	for msg := range msgCh {
		_, err := fmt.Fprintf(os.Stdout, "[%0*d] %s\n", partitionWidth, msg.Partition, msg.Value)
		if err != nil {
			slog.Error("fmt.Fprintf failed", slog.Any("error", err))
		}
	}
}
