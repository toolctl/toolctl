package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func newDiscoverCmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	discoverCmd := &cobra.Command{
		Use:   "discover [TOOL[@VERSION]...] [flags]",
		Short: "Discover new versions of supported tools",
		Example: `  # Discover new versions of all tools
  toolctl discover

  # Discover new versions of a specific tool
  toolctl discover minikube

  # Discover new versions of a specific tool, starting with a specific version
  toolctl discover kubectl@1.20.0`,
		Args: checkArgs(true),
		RunE: newRunDiscover(toolctlWriter, localAPIFS),
	}

	discoverCmd.Flags().StringSlice("arch", []string{"amd64", "arm64"}, "comma-separated list of architectures")
	discoverCmd.Flags().StringSlice("os", []string{"darwin", "linux"}, "comma-separated list of operating systems")

	return discoverCmd
}

func newRunDiscover(toolctlWriter io.Writer, localAPIFS afero.Fs) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// Needs to run with the local API because we need write access
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Local)
		if err != nil {
			return err
		}

		// Get the command line flags
		osArg, err := cmd.Flags().GetStringSlice("os")
		if err != nil {
			return err
		}
		archArg, err := cmd.Flags().GetStringSlice("arch")
		if err != nil {
			return err
		}

		// If no args were specified, discover all tools
		if len(args) == 0 {
			var meta api.Meta
			meta, err = api.GetMeta(toolctlAPI)
			if err != nil {
				return
			}
			args = meta.Tools
		}

		// Convert the arguments to a list of tools
		var allTools []api.Tool
		for _, os := range osArg {
			for _, arch := range archArg {
				tools, err := ArgsToTools(args, os, arch, true)
				if err != nil {
					return err
				}
				allTools = append(allTools, tools...)
			}
		}

		for _, tool := range allTools {
			err = discover(toolctlWriter, toolctlAPI, tool)
			if err != nil {
				return
			}
		}

		return
	}
}

func discover(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, tool api.Tool,
) (err error) {
	// Check if the tool is supported
	toolMeta, err := api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	var version *semver.Version
	if tool.Version != "" {
		if tool.Version == "earliest" {
			var toolPlatformMeta api.ToolPlatformMeta
			toolPlatformMeta, err = api.GetToolPlatformMeta(toolctlAPI, tool)
			if err != nil {
				return
			}
			version = semver.MustParse(toolPlatformMeta.Version.Earliest)
		} else {
			version, err = semver.NewVersion(tool.Version)
			if err != nil {
				return
			}
		}
	} else {
		version, err = setInitialVersion(toolctlAPI, tool)
		if err != nil {
			return
		}
	}

	// Parse the URL template
	funcMap := template.FuncMap{
		"AMD64Bit": func(in string) string {
			return strings.Replace(in, "amd64", "64bit", 1)
		},
		"AMD64Default": func(in string) string {
			return strings.Replace(in, "amd64", "", 1)
		},
		"AMD64X64": func(in string) string {
			return strings.Replace(in, "amd64", "x64", 1)
		},
		"AMD64X86_64": func(in string) string {
			return strings.Replace(in, "amd64", "x86_64", 1)
		},
		"DarwinArchAll": func(in string) string {
			if tool.OS == "darwin" {
				return "all"
			}
			return in
		},
		"DarwinArchUniversal": func(in string) string {
			if tool.OS == "darwin" {
				return "universal"
			}
			return in
		},
		"DarwinExtensionTgz": func(in string) string {
			if tool.OS == "darwin" {
				return ".tgz"
			}
			return in
		},
		"DarwinExtensionZip": func(in string) string {
			if tool.OS == "darwin" {
				return ".zip"
			}
			return in
		},
		"DarwinMacOS": func(in string) string {
			return strings.Replace(in, "darwin", "macOS", 1)
		},
		"LinuxTitle": func(in string) string {
			return strings.Replace(in, "linux", "Linux", 1)
		},
		"LinuxTitleGnu": func(in string) string {
			return strings.Replace(in, "linux", "Linux-gnu", 1)
		},
		"LinuxTitleMusl": func(in string) string {
			return strings.Replace(in, "linux", "Linux-musl", 1)
		},
		"Title": func(in string) string {
			return cases.Title(language.Und, cases.NoLower).String(in)
		},
	}
	downloadURLTemplate, err := template.New("URL").Funcs(funcMap).Parse(toolMeta.DownloadURLTemplate)
	if err != nil {
		return
	}

	tool.Version = version.String()
	err = discoverLoop(
		toolctlWriter, toolctlAPI, toolMeta, tool, downloadURLTemplate,
	)

	return
}

func discoverLoop(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, toolMeta api.ToolMeta,
	tool api.Tool, downloadURLTemplate *template.Template,
) (err error) {
	var (
		componentToIncrement = "patch"
		ignoredVersions      = getIgnoredVersionsMap(toolMeta)
		missCounter          int
		url                  string
		version              = semver.MustParse(tool.Version)
	)

	for {
		var statusCode int
		var skipSleep bool

		tool.Version = version.String()

		// Check if we should ignore this version
		if _, exists := ignoredVersions[tool.Version]; exists {
			fmt.Fprintf(toolctlWriter, "%s %s/%s v%s ignored\n",
				tool.Name, tool.OS, tool.Arch, tool.Version,
			)

			componentToIncrement = "patch"
			missCounter = 0
			skipSleep = true

			version, err = incrementAndSleep(
				version, componentToIncrement, url, skipSleep,
			)
			if err != nil {
				return
			}

			continue
		}

		// Check if we already have this version
		_, err = api.GetToolPlatformVersionMeta(toolctlAPI, tool)
		if err != nil {
			if !errors.Is(err, api.NotFoundError{}) {
				return
			}

			// We don't have the version yet, so we need to check if it's available
			var b bytes.Buffer
			err = downloadURLTemplate.Execute(&b, tool)
			if err != nil {
				return
			}
			url = b.String()

			fmt.Fprintf(toolctlWriter, "%s %s/%s v%s ...\n",
				tool.Name, tool.OS, tool.Arch, tool.Version,
			)
			fmt.Fprintf(toolctlWriter, "URL: %s\n", url)

			statusCode, err = getStatusCode(url)
			if err != nil {
				return
			}

			if statusCode == http.StatusOK {
				err = addNewVersion(toolctlWriter, toolctlAPI, toolMeta, tool, url)
				if err != nil {
					return
				}
				componentToIncrement = "patch"
			} else {
				fmt.Fprintf(toolctlWriter, "HTTP status: %d\n", statusCode)

				missCounter++

				if missCounter > 1 {
					switch componentToIncrement {
					case "patch":
						componentToIncrement = "minor"
					case "minor":
						componentToIncrement = "major"
					case "major":
						return
					}
					missCounter = 0
				}
			}
		} else {
			fmt.Fprintf(toolctlWriter, "%s %s/%s v%s already added\n",
				tool.Name, tool.OS, tool.Arch, tool.Version,
			)
			componentToIncrement = "patch"
			missCounter = 0
			skipSleep = true
		}

		version, err = incrementAndSleep(
			version, componentToIncrement, url, skipSleep,
		)
		if err != nil {
			return
		}
	}
}

func getIgnoredVersionsMap(toolMeta api.ToolMeta) map[string]struct{} {
	ignoredVersions := make(map[string]struct{}, len(toolMeta.IgnoredVersions))
	for _, ignoredVersion := range toolMeta.IgnoredVersions {
		ignoredVersions[ignoredVersion] = struct{}{}
	}
	return ignoredVersions
}

func incrementAndSleep(
	versionToIncrement *semver.Version, componentToIncrement string,
	url string, skipSleep bool,
) (version *semver.Version, err error) {
	version, err = incrementVersion(versionToIncrement, componentToIncrement)
	if err != nil {
		return
	}

	if skipSleep || strings.HasPrefix(url, "http://127.0.0.1:") {
		return
	}

	time.Sleep(500 * time.Millisecond)

	return
}

// addNewVersion adds a new version of a tool to the local API.
func addNewVersion(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, toolMeta api.ToolMeta,
	tool api.Tool, url string,
) (err error) {
	tempDir, err := os.MkdirTemp("", "toolctl-*")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	downloadedToolPath, sha256, err := downloadURL(url, tempDir)
	if err != nil {
		return
	}

	// Check the version, if we can run the tool binary
	if tool.OS == runtime.GOOS && tool.Arch == runtime.GOARCH {
		// Extract the downloaded tool
		var extractedToolPath string
		extractedToolPath, err = extractDownloadedTool(tool, downloadedToolPath)
		if err != nil {
			return
		}

		// Check the version
		var toolBinaryVersion *semver.Version
		toolBinaryVersion, err = getToolBinaryVersion(
			extractedToolPath, toolMeta.VersionArgs,
		)
		if err != nil {
			return
		}
		if !toolBinaryVersion.Equal(semver.MustParse(tool.Version)) {
			err = fmt.Errorf(
				"version mismatch: expected %s, got %s",
				tool.Version, toolBinaryVersion,
			)
			return
		}
	}

	fmt.Fprintln(toolctlWriter, "SHA256:", sha256)

	// Save the tool platform version metadata
	toolPlatformVersionMeta := api.ToolPlatformVersionMeta{
		URL:    url,
		SHA256: sha256,
	}
	err = api.SaveToolPlatformVersionMeta(toolctlAPI, tool, toolPlatformVersionMeta)
	if err != nil {
		return
	}

	// Update the tool platform metadata
	err = updateToolPlatformMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	return
}

func updateToolPlatformMeta(toolctlAPI api.ToolctlAPI, tool api.Tool) (err error) {
	var toolPlatformMeta api.ToolPlatformMeta
	toolPlatformMeta, err = api.GetToolPlatformMeta(toolctlAPI, tool)
	if err != nil {
		if !errors.Is(err, api.NotFoundError{}) {
			return
		}

		toolPlatformMeta = api.ToolPlatformMeta{
			Version: api.ToolPlatformMetaVersion{
				Earliest: "42.0.0",
				Latest:   "0.0.0",
			},
		}
		err = api.SaveToolPlatformMeta(toolctlAPI, tool, toolPlatformMeta)
		if err != nil {
			return
		}
	}

	var earliestVersion *semver.Version
	earliestVersion, err = semver.NewVersion(toolPlatformMeta.Version.Earliest)
	if err != nil {
		earliestVersion = semver.MustParse("42.0.0")
	}

	version := semver.MustParse(tool.Version)
	if version.LessThan(earliestVersion) {
		toolPlatformMeta.Version.Earliest = version.String()
	}

	var latestVersion *semver.Version
	latestVersion, err = semver.NewVersion(toolPlatformMeta.Version.Latest)
	if err != nil {
		latestVersion = semver.MustParse("0.0.0")
	}
	if version.GreaterThan(latestVersion) {
		toolPlatformMeta.Version.Latest = version.String()
	}

	err = api.SaveToolPlatformMeta(toolctlAPI, tool, toolPlatformMeta)
	if err != nil {
		return
	}

	return
}

func setInitialVersion(toolctlAPI api.ToolctlAPI, noa api.Tool) (version *semver.Version, err error) {
	version, err = api.GetLatestVersion(toolctlAPI, noa)
	if err != nil {
		if !errors.Is(err, api.NotFoundError{}) {
			return
		}
		version = semver.MustParse("0.0.0")
	}
	version, err = incrementVersion(version, "patch")
	if err != nil {
		return
	}

	return
}

func getStatusCode(url string) (statusCode int, err error) {
	resp, err := http.Head(url)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	return
}

func incrementVersion(version *semver.Version, component string) (*semver.Version, error) {
	var incrementedVersion semver.Version

	switch component {
	case "major":
		if version.Patch() == 0 {
			incrementedVersion = version.IncPatch()
		} else {
			incrementedVersion = version.IncMajor()
		}
	case "minor":
		if version.Patch() == 0 {
			incrementedVersion = version.IncPatch()
		} else {
			incrementedVersion = version.IncMinor()
		}
	case "patch":
		incrementedVersion = version.IncPatch()
	default:
		return nil, fmt.Errorf("invalid version component: %s", component)
	}

	return &incrementedVersion, nil
}
