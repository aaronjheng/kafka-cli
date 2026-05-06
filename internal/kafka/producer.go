package kafka

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

const keyValueParts = 2

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
		}
	}

	return nil
}
