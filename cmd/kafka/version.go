package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var embeddedVersion string

var version = strings.TrimRight(embeddedVersion, "\n")

func buildCommit() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}

	return ""
}

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(_ *cobra.Command, _ []string) {
			commit := buildCommit()
			if commit == "" {
				_, _ = fmt.Fprintf(os.Stdout, "%s (%s %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)

				return
			}

			_, _ = fmt.Fprintf(
				os.Stdout,
				"%s (%s %s/%s) (commit: %s)\n",
				version,
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
				commit,
			)
		},
	}

	return cmd
}
