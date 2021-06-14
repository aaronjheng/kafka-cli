package main

import (
	"bufio"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"
)

var producerCmd = &cobra.Command{
	Use:   "producer",
	Short: "producer",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var producerConsoleCmd = &cobra.Command{
	Use: "console",
	Run: func(cmd *cobra.Command, args []string) {
		producer, err := newSyncProducer()
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := producer.Close(); err != nil {
				log.Panicln(err)
			}
		}()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(scanner.Text()),
			}
			_, _, err := producer.SendMessage(msg)
			if err != nil {
				log.Panicln(err)
			}
		}
	},
}
