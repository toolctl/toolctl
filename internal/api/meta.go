//nolint:revive // package name is intentionally concise
package api

import (
	"bytes"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Meta contains metadata for the toolctl API.
type Meta struct {
	Tools []string
}

// GetMeta returns the metadata for the toolctl API.
func GetMeta(toolctlAPI ToolctlAPI) (meta Meta, err error) {
	var found bool
	var metaBytes []byte
	found, metaBytes, err = toolctlAPI.GetContents("meta.yaml")
	if err != nil {
		return
	}
	if !found {
		err = fmt.Errorf("global metadata %w", NotFoundError{})
		return
	}

	err = yaml.Unmarshal(metaBytes, &meta)
	return
}

// SaveMeta saves the metadata for the toolctl API.
func SaveMeta(toolctlAPI ToolctlAPI, meta Meta) (err error) {
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

	err = toolctlAPI.SaveContents("meta.yaml", yamlBuffer.Bytes())
	return
}

// ToolMeta contains metadata for a tool.
type ToolMeta struct {
	Description         string
	DownloadURLTemplate string `yaml:"downloadURLTemplate"`
	Homepage            string
	IgnoredVersions     []string `yaml:"ignoredVersions"`
	VersionArgs         []string `yaml:"versionArgs"`
}

// GetToolMeta returns the metadata for the given tool.
func GetToolMeta(toolctlAPI ToolctlAPI, tool Tool) (meta ToolMeta, err error) {
	var found bool
	var metaBytes []byte
	found, metaBytes, err = toolctlAPI.GetContents(tool.Name + "/meta.yaml")
	if err != nil {
		return
	}
	if !found {
		err = fmt.Errorf("%s %w", tool.Name, NotFoundError{})
		return
	}

	err = yaml.Unmarshal(metaBytes, &meta)
	return
}

// ToolPlatformMeta contains metadata for a given tool and platform.
type ToolPlatformMeta struct {
	Version ToolPlatformMetaVersion
}

// ToolPlatformMetaVersion contains version metadata for a given tool and platform.
type ToolPlatformMetaVersion struct {
	Earliest string
	Latest   string
}

// GetToolPlatformMeta returns the metadata for the given tool and platform.
func GetToolPlatformMeta(toolctlAPI ToolctlAPI, tool Tool) (meta ToolPlatformMeta, err error) {
	var found bool
	var metaBytes []byte
	found, metaBytes, err = toolctlAPI.GetContents(tool.Name + "/" + tool.OS + "-" + tool.Arch + "/meta.yaml")
	if err != nil {
		return
	}
	if !found {
		err = fmt.Errorf("%s %w", tool.Name, NotFoundError{})
		return
	}

	err = yaml.Unmarshal([]byte(metaBytes), &meta)
	return
}

// SaveToolPlatformMeta saves the metadata for the given tool and platform.
func SaveToolPlatformMeta(toolctlAPI ToolctlAPI, tool Tool, meta ToolPlatformMeta) (err error) {
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

	err = toolctlAPI.SaveContents(
		filepath.Join(tool.Name, tool.OS+"-"+tool.Arch, "meta.yaml"),
		yamlBuffer.Bytes(),
	)
	return
}

// ToolPlatformVersionMeta contains metadata for a given tool version and platform.
type ToolPlatformVersionMeta struct {
	URL    string
	SHA256 string
}

// GetToolPlatformVersionMeta returns the metadata for the given tool version and platform.
func GetToolPlatformVersionMeta(toolctlAPI ToolctlAPI, tool Tool) (meta ToolPlatformVersionMeta, err error) {
	var found bool
	var metaBytes []byte
	found, metaBytes, err = toolctlAPI.GetContents(
		tool.Name + "/" + tool.OS + "-" + tool.Arch + "/" + tool.Version + ".yaml",
	)
	if err != nil {
		return
	}
	if !found {
		err = fmt.Errorf("%s v%s %w", tool.Name, tool.Version, NotFoundError{})
		return
	}

	err = yaml.Unmarshal([]byte(metaBytes), &meta)
	return
}

// SaveToolPlatformVersionMeta saves version metadata for a tool to the API.
func SaveToolPlatformVersionMeta(toolctlAPI ToolctlAPI, tool Tool, meta ToolPlatformVersionMeta) (err error) {
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

	err = toolctlAPI.SaveContents(
		filepath.Join(tool.Name, tool.OS+"-"+tool.Arch, tool.Version+".yaml"),
		yamlBuffer.Bytes(),
	)
	return
}
