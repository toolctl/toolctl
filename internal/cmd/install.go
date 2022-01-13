package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
	"github.com/toolctl/toolctl/internal/utils"
	"golang.org/x/sys/unix"
)

func newInstallCmd(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) *cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "install TOOL[@VERSION]... [flags]",
		Short: "Install tools",
		Example: `  # Install the latest version of a tool
  toolctl install minikube

  # Install a specified version of a tool
  toolctl install kubectl@1.20.13

  # Install multiple tools
  toolctl install gh k9s`,
		Args: checkArgs(false),
		RunE: newRunInstall(toolctlWriter, localAPIFS),
	}
	return installCmd
}

func newRunInstall(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Remote)
		if err != nil {
			return err
		}

		allTools, err := ArgsToTools(args, runtime.GOOS, runtime.GOARCH, true)
		if err != nil {
			return err
		}

		installDir, err := checkInstallDir(toolctlWriter, "install", args)
		if err != nil {
			return
		}

		for _, tool := range allTools {
			err = install(toolctlWriter, toolctlAPI, installDir, tool, allTools)
			if err != nil {
				return
			}
		}

		return
	}
}

func checkInstallDir(
	toolctlWriter io.Writer, installOrUpgrade string, args []string,
) (installDir string, err error) {
	installDir, err = utils.RequireConfigString("InstallDir")
	if err != nil {
		return
	}
	_, err = os.Stat(installDir)
	if err != nil {
		err = fmt.Errorf(
			"install directory %s does not exist",
			wrapInQuotesIfContainsSpace(installDir),
		)
		return
	}

	if unix.Access(installDir, unix.W_OK) != nil {
		var currentUser *user.User
		currentUser, err = user.Current()
		if err != nil {
			return
		}

		joinedArgs := strings.Join(args, " ")
		if joinedArgs != "" {
			joinedArgs = " " + joinedArgs
		}
		err = fmt.Errorf("%s is not writable by user %s, try running:\n  sudo toolctl %s%s",
			wrapInQuotesIfContainsSpace(installDir), currentUser.Username,
			installOrUpgrade, joinedArgs,
		)

		return
	}

	var installDirInPath bool
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, ":")
	for _, path := range paths {
		if path == installDir {
			installDirInPath = true
			break
		}
	}
	if !installDirInPath {
		fmt.Fprintf(
			toolctlWriter,
			"🚨 %s is not in $PATH\n",
			wrapInQuotesIfContainsSpace(installDir),
		)
	}

	return
}

func install(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, installDir string,
	tool api.Tool, allTools []api.Tool,
) (err error) {
	// Check if the tool is supported
	toolMeta, err := api.GetToolMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	// Check if a version has been specified
	latestVersion, err := api.GetLatestVersion(toolctlAPI, tool)
	if err != nil {
		return
	}
	if tool.Version == "" {
		tool.Version = latestVersion.String()
	}

	// Check if the tool is already installed
	installedToolPath, err := which(tool.Name)
	if err != nil {
		return
	}

	if installedToolPath != "" {
		err = infoPrintInstalledVersion(
			installedToolPath, toolMeta, toolctlWriter, tool, allTools, latestVersion,
		)
		return
	}

	fmt.Fprintln(
		toolctlWriter,
		prependToolName(tool, allTools, fmt.Sprintf(
			"👷 Installing v%s ...", tool.Version),
		),
	)

	// Download the tool
	tempDir, err := ioutil.TempDir("", "toolctl-*")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	downloadedToolPath, err := downloadTool(toolctlAPI, tool, tempDir)
	if err != nil {
		return
	}

	// Extract the tool
	extractedToolPath, err := extractDownloadedTool(tool, downloadedToolPath)
	if err != nil {
		return
	}

	// Install the tool
	installPath := filepath.Join(installDir, tool.Name)
	err = os.Rename(extractedToolPath, installPath)
	if err != nil {
		return
	}

	installedVersion, err := getToolBinaryVersion(
		installPath, toolMeta.VersionArgs,
	)
	if err != nil {
		return
	}

	if !installedVersion.Equal(semver.MustParse(tool.Version)) {
		err = fmt.Errorf(
			"installation failed: expected v%s, but installed binary reported v%s",
			tool.Version, installedVersion.String(),
		)
		return
	}

	fmt.Fprintln(
		toolctlWriter,
		prependToolName(tool, allTools, "🎉 Successfully installed"),
	)

	return
}

func infoPrintInstalledVersion(
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
			prependToolName(
				tool, allTools, "🤷 Unknown version is already installed",
			),
		)
		fmt.Fprintln(
			toolctlWriter,
			prependToolName(
				tool, allTools, "💁 For more details, run: toolctl info "+tool.Name,
			),
		)
		err = nil
		return
	}

	installedVersionString := installedVersion.String()
	if installedVersion.Equal(latestVersion) {
		installedVersionString += " (the latest version)"
	}
	fmt.Fprintln(
		toolctlWriter,
		prependToolName(tool, allTools, fmt.Sprintf(
			"🤷 v%s is already installed", installedVersionString),
		),
	)

	fmt.Fprintln(
		toolctlWriter,
		prependToolName(
			tool, allTools, "💁 For more details, run: toolctl info "+tool.Name,
		),
	)

	return
}

type binaryLocatedError struct{}

func (b binaryLocatedError) Error() string {
	return "binary located"
}

func locateExtractedBinary(dir string, tool api.Tool) (
	extractedBinaryPath string, err error,
) {
	err = filepath.WalkDir(dir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if filepath.Base(path) == tool.Name ||
				filepath.Base(path) == tool.Name+"-"+runtime.GOOS+"-"+runtime.GOARCH ||
				filepath.Base(path) == tool.Name+"_"+runtime.GOOS+"_"+runtime.GOARCH {
				extractedBinaryPath = path
				return fmt.Errorf("%w", binaryLocatedError{})
			}

			return nil
		},
	)

	if err != nil {
		if !errors.Is(err, binaryLocatedError{}) {
			return
		}
		err = nil
	}

	if extractedBinaryPath == "" {
		err = fmt.Errorf(
			"failed to locate extracted binary for %s",
			tool.Name,
		)
	}

	return
}

// downloadTool gets the download URL for the specified tool and
// downloads it to the specified directory.
func downloadTool(
	toolctlAPI api.ToolctlAPI, tool api.Tool, dir string,
) (downloadedToolPath string, err error) {
	meta, err := api.GetToolPlatformVersionMeta(toolctlAPI, tool)
	if err != nil {
		return
	}

	var sha256 string
	downloadedToolPath, sha256, err = downloadURL(meta.URL, dir)
	if err != nil {
		return
	}

	if sha256 != meta.SHA256 {
		err = fmt.Errorf(
			"SHA256 hash mismatch, wanted %s, got %s",
			meta.SHA256, sha256,
		)
		return
	}

	return
}
