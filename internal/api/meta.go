package api

import (
	"bytes"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Meta struct {
	Tools []string
}

// GetToolMeta returns the metadata for the given tool
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

type ToolMeta struct {
	Description         string
	DownloadURLTemplate string `yaml:"downloadURLTemplate"`
	Homepage            string
	VersionArgs         []string `yaml:"versionArgs"`
}

// GetToolMeta returns the metadata for the given tool
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

type ToolPlatformMeta struct {
	Version ToolPlatformMetaVersion
}
type ToolPlatformMetaVersion struct {
	Earliest string
	Latest   string
}

// GetToolPlatformMeta returns the metadata for the given tool, OS and architecture
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
	if err != nil {
		return
	}

	return
}

type ToolPlatformVersionMeta struct {
	URL    string
	SHA256 string
}

func GetToolPlatformVersionMeta(toolctlAPI ToolctlAPI, tool Tool) (meta ToolPlatformVersionMeta, err error) {
	var found bool
	var metaBytes []byte
	found, metaBytes, err = toolctlAPI.GetContents(tool.Name + "/" + tool.OS + "-" + tool.Arch + "/" + tool.Version + ".yaml")
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
