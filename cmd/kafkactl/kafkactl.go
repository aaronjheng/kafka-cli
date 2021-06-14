package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafkactl/pkg/config"
)

var cfg *config.Config
var profile string

var rootCmd = &cobra.Command{
	Use: "kafkactl",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "")

	rootCmd.AddCommand(topicCmd)
	topicCmd.AddCommand(topicListCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
