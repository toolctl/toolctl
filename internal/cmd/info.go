package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

func newInfoCmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	var infoCmd = &cobra.Command{
		Use:   "info [TOOL...]",
		Short: "Get information about tools",
		Args:  checkArgs(true),
		Example: `  # Get information about installed tools
  toolctl info

  # Get information about a specific tool
  toolctl info kubectl

  # Get information about multiple tools
  toolctl info gh k9s`,
		RunE: newRunInfo(toolctlWriter, localAPIFS),
	}
	return infoCmd
}

func newRunInfo(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) func(*cobra.Command, []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Remote)
		if err != nil {
			return err
		}

		// If no args were specified, upgrade all installed tools
		if len(args) == 0 {
			// Get the list of all tools
			var meta api.Meta
			meta, err = api.GetMeta(toolctlAPI)
			if err != nil {
				return
			}

			// Check which tools are installed
			var installedToolNames []string
			for _, toolName := range meta.Tools {
				var installed bool
				installed, err = isToolInstalled(toolName)
				if err != nil {
					return
				}
				if installed {
					installedToolNames = append(installedToolNames, toolName)
				}
			}

			if len(installedToolNames) == 0 {
				err = fmt.Errorf("no supported tools installed")
				return
			}

			args = installedToolNames
		}

		allTools, err := ArgsToTools(args, runtime.GOOS, runtime.GOARCH, false)
		if err != nil {
			return fmt.Errorf(
				"%w, try this instead:\n  toolctl info %s",
				err, strings.Join(stripVersionsFromArgs(args), " "),
			)
		}

		for _, tool := range allTools {
			err = info(toolctlWriter, toolctlAPI, tool, allTools)
			if err != nil {
				return
			}
		}

		return
	}
}

func info(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, tool api.Tool,
	allTools []api.Tool,
) (err error) {
	// Check if the tool is supported
	toolMeta, err := api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	var latestVersion *semver.Version
	latestVersion, err = api.GetLatestVersion(toolctlAPI, tool)
	if err != nil {
		if errors.Is(err, api.NotFoundError{}) {
			err = fmt.Errorf("%s not supported on this platform", tool.Name)
		}
		return
	}

	fmt.Fprintln(
		toolctlWriter,
		prependToolName(tool, allTools, fmt.Sprintf("✨ %s v%s: %s",
			tool.Name, latestVersion.String(), toolMeta.Description),
		),
	)

	// Check if the tool is already installed
	installedToolPath, err := which(tool.Name)
	if err != nil {
		return
	}
	if installedToolPath == "" {
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, fmt.Sprintf("🏠 %s",
				toolMeta.Homepage),
			),
		)
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, "❌ Not installed"),
		)
		return
	}

	err = installPrintInstalledVersion(
		installedToolPath, toolMeta, toolctlWriter, tool, allTools, latestVersion,
	)
	if err != nil {
		return
	}

	// Check if the tool path is a symlink
	var fi fs.FileInfo
	fi, err = os.Lstat(installedToolPath)
	if err != nil {
		return
	}
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		var symlink string
		symlink, err = filepath.EvalSymlinks(installedToolPath)
		if err != nil {
			return
		}
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools,
				"🔗 Symlinked from", wrapInQuotesIfContainsSpace(symlink),
			),
		)
	}

	return
}

func installPrintInstalledVersion(
	installedToolPath string, toolMeta api.ToolMeta, toolctlWriter io.Writer,
	tool api.Tool, allTools []api.Tool, latestVersion *semver.Version,
) (err error) {
	var installedVersion *semver.Version
	installedVersion, err = getToolBinaryVersion(
		installedToolPath, toolMeta.VersionArgs,
	)
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return
		}

		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, err.Error()),
		)
		err = nil
		return
	}

	if installedVersion.Equal(latestVersion) {
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, fmt.Sprintf(
				"✅ %s v%s is installed at %s",
				tool.Name, installedVersion.String(),
				wrapInQuotesIfContainsSpace(installedToolPath)),
			),
		)
	} else {
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, fmt.Sprintf(
				"🔄 %s v%s is installed at %s",
				tool.Name, installedVersion.String(),
				wrapInQuotesIfContainsSpace(installedToolPath)),
			),
		)
	}

	return
}
