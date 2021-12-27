package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

func newUpgradeCmd(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) *cobra.Command {
	var upgradeCmd = &cobra.Command{
		Use:   "upgrade TOOL... [flags]",
		Short: "Upgrade one or more tools",
		Example: `  # Upgrade a tool
  toolctl upgrade minikube

  # Upgrade multiple tools
  toolctl upgrade kustomize k9s`,
		Args: checkArgs(false),
		RunE: newRunUpgrade(toolctlWriter, localAPIFS),
	}
	return upgradeCmd
}

func newRunUpgrade(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Remote)
		if err != nil {
			return err
		}

		allTools, err := ArgsToTools(args, runtime.GOOS, runtime.GOARCH, false)
		if err != nil {
			return fmt.Errorf(
				"%w, try this instead:\n  toolctl upgrade %s",
				err, strings.Join(stripVersionsFromArgs(args), " "),
			)
		}

		installDir, err := checkInstallDir(toolctlWriter, allTools, "upgrade")
		if err != nil {
			return
		}

		for _, tool := range allTools {
			err = upgrade(toolctlWriter, toolctlAPI, installDir, tool, allTools)
			if err != nil {
				return
			}
		}

		return
	}
}

func upgrade(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, installDir string,
	tool api.Tool, allTools []api.Tool,
) (err error) {
	// Check if the tool is supported
	_, err = api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	// Check if the tool is installed
	installedToolPath, err := which(tool.Name)
	if err != nil {
		return
	}
	if installedToolPath == "" {
		err = fmt.Errorf(
			"%s is not installed", tool.Name,
		)
		return
	}

	// Check if the tool is installed in a different directory
	if filepath.Dir(installedToolPath) != installDir {
		err = fmt.Errorf(
			"aborting: %s is currently installed in %s, not in %s",
			tool.Name, filepath.Dir(installedToolPath), installDir,
		)
		return
	}

	// Get the latest version
	latestVersion, err := api.GetLatestVersion(toolctlAPI, tool)
	if err != nil {
		return
	}

	// Get the installed version
	toolMeta, err := api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}
	installedVersion, err := getToolBinaryVersion(
		installedToolPath, toolMeta.VersionArgs,
	)

	// Check if the installed version is newer than the latest version
	if installedVersion.GreaterThan(latestVersion) {
		err = fmt.Errorf(
			"%s is already at v%s, but the latest version is v%s",
			tool.Name, installedVersion, latestVersion,
		)
		return
	}

	// Check if the installed version is the latest version
	if installedVersion.Equal(latestVersion) {
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(tool, allTools, "✅ already up-to-date"),
		)
		return
	}

	// Check if the installed tool is symlinked
	var fi fs.FileInfo
	fi, err = os.Lstat(installedToolPath)
	if err != nil {
		return
	}
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		var symlinkPath string
		symlinkPath, err = filepath.EvalSymlinks(installedToolPath)
		if err != nil {
			return
		}
		err = fmt.Errorf(
			"aborting: %s is symlinked from %s",
			wrapInQuotesIfContainsSpace(installedToolPath),
			wrapInQuotesIfContainsSpace(symlinkPath),
		)
		return
	}

	// Start the upgrade
	fmt.Fprintln(
		toolctlWriter, prependToolName(
			tool, allTools, fmt.Sprintf(
				"👷 Upgrading from v%s to v%s ...", installedVersion, latestVersion,
			),
		),
	)

	// Remove the installed tool
	fmt.Fprintln(
		toolctlWriter, prependToolName(
			tool, allTools, fmt.Sprintf(
				"👷 Removing v%s ...", installedVersion,
			),
		),
	)
	err = os.Remove(installedToolPath)
	if err != nil {
		return
	}

	// Install the latest version
	err = install(toolctlWriter, toolctlAPI, installDir, tool, allTools)
	if err != nil {
		return
	}

	return
}
