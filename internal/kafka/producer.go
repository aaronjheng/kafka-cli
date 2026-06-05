package kafka

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

const (
	keyValueParts        = 2
	scannerMaxBufferSize = 1024 * 1024
)

func RunTopicProduce(clusterCfg *Config, topic string, keySeparator string) error {
	producer, err := NewSyncProducer(clusterCfg)
	if err != nil {
		return fmt.Errorf("kafka.NewSyncProducer error: %w", err)
	}

	defer func() {
		err := producer.Close()
		if err != nil {
			slog.Error("producer.Close failed", slog.Any("error", err))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), scannerMaxBufferSize)

	var sendErr error

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(line),
		}

		if keySeparator != "" {
			parts := strings.SplitN(line, keySeparator, keyValueParts)
			if len(parts) == keyValueParts {
				msg.Key = sarama.StringEncoder(parts[0])
				msg.Value = sarama.StringEncoder(parts[1])
			}
		}

		_, _, err := producer.SendMessage(msg)
		if err != nil {
			slog.Error("producer.SendMessage failed", slog.Any("error", err))
			sendErr = errors.Join(sendErr, err)
		}
	}

	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("scan stdin error: %w", err)
	}

	if sendErr != nil {
		return fmt.Errorf("producer.SendMessage error: %w", sendErr)
	}

	return nil
}
