package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var embeddedVersion string

var version = strings.TrimRight(embeddedVersion, "\n")

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "version",
		Run: func(_ *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(os.Stdout, "%s (%s %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	}

	return cmd
}
