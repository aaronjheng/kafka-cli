package main

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var embeddedVersion string

const splitNParts = 2

var (
	version                  = strings.TrimRight(embeddedVersion, "\n")
	pseudoVersionCommitRegex = regexp.MustCompile(`-[0-9]{14}-([0-9a-f]{12,40})$`)
)

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

	moduleVersion := strings.SplitN(buildInfo.Main.Version, "+", splitNParts)[0]
	matches := pseudoVersionCommitRegex.FindStringSubmatch(moduleVersion)

	if len(matches) == splitNParts {
		return matches[1]
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
