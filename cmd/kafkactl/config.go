package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var configCmd = &cobra.Command{
	Use: "config",
}

var configCatCmd = &cobra.Command{
	Use: "cat",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPathname := cfg.Filepath()
		fmt.Printf("# %s\n", cfgPathname)
		f, err := os.Open(cfgPathname)
		if err != nil {
			logger.Error("Open file error", zap.Error(err))
		}
		defer f.Close()

		io.Copy(os.Stdout, f)

		return nil
	},
}
