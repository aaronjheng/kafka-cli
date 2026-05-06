package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

const (
	splitNParts         = 2
	expectedMatchGroups = 2
)

var pseudoVersionCommitRegex = regexp.MustCompile(`-[0-9]{14}-([0-9a-f]{12,40})$`)

func BuildCommit() string {
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
	if len(matches) == expectedMatchGroups {
		return matches[1]
	}

	return ""
}
