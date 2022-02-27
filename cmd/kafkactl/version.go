package main

import (
	_ "embed"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}
