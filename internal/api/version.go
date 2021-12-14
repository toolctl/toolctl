package api

import (
	"bytes"
	"path/filepath"

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v3"
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

func SaveVersion(toolctlAPI ToolctlAPI, tool Tool, meta ToolPlatformVersionMeta) (err error) {
	yamlBuffer := &bytes.Buffer{}

	yamlEncoder := yaml.NewEncoder(yamlBuffer)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(meta)
	if err != nil {
		return
	}
	err = yamlEncoder.Close()
	if err != nil {
		return
	}

	relativePath := filepath.Join(
		tool.Name, tool.OS+"-"+tool.Arch, tool.Version+".yaml",
	)

	err = toolctlAPI.SaveContents(relativePath, yamlBuffer.Bytes())
	if err != nil {
		return
	}

	return
}
