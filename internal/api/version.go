package api

import (
	"github.com/Masterminds/semver"
)

// GetLatestVersion returns the latest version for the given tool, OS and arch
func GetLatestVersion(toolctlAPI ToolctlAPI, tool Tool) (version *semver.Version, err error) {
	toolPlatformMeta, err := GetToolPlatformMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	version, err = semver.NewVersion(toolPlatformMeta.Version.Latest)
	if err != nil {
		return
	}

	return
}
