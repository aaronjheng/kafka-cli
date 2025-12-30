package main

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var version string

func init() {
	version = strings.TrimRight(version, "\n")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}
