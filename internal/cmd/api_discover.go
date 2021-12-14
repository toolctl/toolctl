package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

func newDiscoverCmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	discoverCmd := &cobra.Command{
		Use:   "discover TOOL[@VERSION]... [flags]",
		Short: "Discover new versions of one or more tools",
		Example: `  # Discover new versions of kubectl
  toolctl discover kubectl

  # Discover new versions of kubectl, starting with v1.20.0
  toolctl discover kubectl@1.20.0`,
		Args: checkArgs(),
		RunE: newRunDiscover(toolctlWriter, localAPIFS),
	}
	return discoverCmd
}

func newRunDiscover(toolctlWriter io.Writer, localAPIFS afero.Fs) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// Needs to run with the local API because we need write access
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Local)
		if err != nil {
			return err
		}

		allTools, err := ArgsToTools(args, true)
		if err != nil {
			return err
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

func discover(toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, tool api.Tool) (err error) {
	// Check if the tool is supported
	toolMeta, err := api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	var version *semver.Version
	if tool.Version != "" {
		version = semver.MustParse(tool.Version)
	} else {
		version, err = setInitialVersion(toolctlAPI, tool)
		if err != nil {
			return
		}
	}

	// Parse the URL template
	funcMap := template.FuncMap{
		"Title": strings.Title,
		"X86_64": func(in string) string {
			return strings.Replace(in, "amd64", "x86_64", 1)
		},
	}

	downloadURLTemplate, err := template.New("URL").Funcs(funcMap).Parse(toolMeta.DownloadURLTemplate)
	if err != nil {
		return
	}

	// Discover new versions
	var (
		url                  string
		componentToIncrement string = "patch"
		missCounter          int    = 0
	)

	for {
		var statusCode int

		// Check if we already have the version
		tool.Version = version.String()
		_, err = api.GetToolPlatformVersionMeta(toolctlAPI, tool)
		if err != nil {
			if !errors.Is(err, api.NotFoundError{}) {
				return
			}
		} else {
			fmt.Fprintf(toolctlWriter, "%s v%s already added\n",
				tool.Name, tool.Version,
			)
			// Hacky McHackface
			statusCode = http.StatusFound
			componentToIncrement = "patch"
		}

		if statusCode != http.StatusFound {
			tool.Version = version.String()

			var b bytes.Buffer
			err = downloadURLTemplate.Execute(&b, tool)
			if err != nil {
				return
			}
			url = b.String()

			fmt.Fprintf(toolctlWriter, "%s v%s ...\n", tool.Name, tool.Version)
			fmt.Fprintf(toolctlWriter, "URL: %s\n", url)

			statusCode, err = getStatusCode(url)
			if err != nil {
				return
			}

			if statusCode == http.StatusOK {
				// Download the binary and calculate the SHA256
				var resp *http.Response
				resp, err = http.Get(url)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				var sha256 string
				sha256, err = CalculateSHA256(resp.Body)
				if err != nil {
					return
				}
				fmt.Fprintln(toolctlWriter, "SHA256:", sha256)

				// Save the tool platform version metadata
				toolPlatformVersionMeta := api.ToolPlatformVersionMeta{
					URL:    url,
					SHA256: sha256,
				}
				err = api.SaveVersion(toolctlAPI, tool, toolPlatformVersionMeta)
				if err != nil {
					return
				}

				// Update the tool platform metadata
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

				componentToIncrement = "patch"
			} else {
				fmt.Fprintf(toolctlWriter, "HTTP status: %d\n", statusCode)

				missCounter++

				if missCounter > 1 {
					if componentToIncrement == "patch" {
						componentToIncrement = "minor"
					} else if componentToIncrement == "minor" {
						componentToIncrement = "major"
					} else if componentToIncrement == "major" {
						return nil
					}
					missCounter = 0
				}
			}
		}

		version, err = incrementVersion(version, componentToIncrement)
		if err != nil {
			return err
		}

		// If the URL starts with "http://127.0.0.1:", we don't need to throttle
		if statusCode != http.StatusFound && !strings.HasPrefix(url, "http://127.0.0.1:") {
			time.Sleep(1 * time.Second)
		}
	}
}

func setInitialVersion(toolctlAPI api.ToolctlAPI, noa api.Tool) (version *semver.Version, err error) {
	version, err = api.GetLatestVersion(toolctlAPI, noa)
	if err != nil {
		if !errors.Is(err, api.NotFoundError{}) {
			return
		} else {
			version = semver.MustParse("0.0.0")
		}
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
		incrementedVersion = version.IncMajor()
	case "minor":
		incrementedVersion = version.IncMinor()
	case "patch":
		incrementedVersion = version.IncPatch()
	default:
		return nil, fmt.Errorf("invalid version component: %s", component)
	}

	return &incrementedVersion, nil
}
