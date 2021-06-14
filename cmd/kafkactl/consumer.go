package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/Shopify/sarama"
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
	Run: func(cmd *cobra.Command, args []string) {
		consumer, err := newConsumer()
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := consumer.Close(); err != nil {
				log.Panicln(err)
			}
		}()

		// Partition flag not specified
		var partitions []int32
		if partition == -1 {
			var err error
			partitions, err = consumer.Partitions(topic)
			if err != nil {
				log.Fatal(err)
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
					log.Fatal(err)
				}

				for {
					select {
					case msg := <-partitionConsumer.Messages():
						msgCh <- string(msg.Value)
					}
				}

			}(partition)
		}

		go func() {
			for msg := range msgCh {
				fmt.Println(msg)
			}
		}()

		wg.Wait()
	},
}
