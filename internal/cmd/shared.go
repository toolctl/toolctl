package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

// ArgToTool converts a command line argument to a tool.
func ArgToTool(arg string, os string, arch string, versionAllowed bool) (tool api.Tool, err error) {
	splitArg := strings.SplitN(arg, "@", 2)
	tool = api.Tool{
		Name: splitArg[0],
		OS:   os,
		Arch: arch,
	}
	if len(splitArg) == 2 {
		if !versionAllowed {
			err = fmt.Errorf("please don't specify a tool version")
			return
		}
		tool.Version = splitArg[1]
	}
	return
}

// ArgsToTools converts a list of command line arguments to a list of tools.
func ArgsToTools(
	args []string, os string, arch string, versionAllowed bool,
) ([]api.Tool, error) {
	var tools []api.Tool

	for _, arg := range args {
		tool, err := ArgToTool(arg, os, arch, versionAllowed)
		if err != nil {
			return []api.Tool{}, err
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// CalculateSHA256 calculates the SHA256 hash of an io.Reader.
func CalculateSHA256(body io.Reader) (sha string, err error) {
	hash := sha256.New()
	_, err = io.Copy(hash, body)
	if err != nil {
		return
	}
	sha = fmt.Sprintf("%x", hash.Sum(nil))
	return
}

func checkArgs(worksWithoutArgs bool) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) (err error) {
		if !worksWithoutArgs && len(args) == 0 {
			err = fmt.Errorf("no tool specified")
		}
		return
	}
}

// downloadURL downloads the specified URL to the specified directory.
func downloadURL(url string, dir string) (
	downloadedFilePath string, sha256 string, err error,
) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	downloadedFilePath = filepath.Join(dir, path.Base(url))

	// Create the file
	downloadedFile, err := os.Create(downloadedFilePath)
	if err != nil {
		return
	}
	_, err = io.Copy(downloadedFile, resp.Body)
	if err != nil {
		return
	}
	err = downloadedFile.Close()
	if err != nil {
		return
	}

	// Calculate the SHA256 hash
	downloadedFile, err = os.Open(downloadedFilePath)
	if err != nil {
		return
	}
	sha256, err = CalculateSHA256(downloadedFile)
	if err != nil {
		return
	}
	err = downloadedFile.Close()

	return
}

// extractDownloadedTool extracts the downloaded tool.
func extractDownloadedTool(tool api.Tool, downloadedToolPath string) (
	extractedToolPath string, err error,
) {
	dir := filepath.Dir(downloadedToolPath)

	// Extract the tool if it's a .tar.gz file
	if strings.HasSuffix(downloadedToolPath, ".tar.gz") ||
		strings.HasSuffix(downloadedToolPath, ".tgz") ||
		strings.HasSuffix(downloadedToolPath, ".zip") {
		// Extract the downloaded file
		err = archiver.Unarchive(downloadedToolPath, dir)
		if err != nil {
			return
		}
		// Locate the extracted binary
		extractedToolPath, err = locateExtractedBinary(dir, tool)
		if err != nil {
			return
		}
	} else {
		extractedToolPath = downloadedToolPath
	}

	// Make sure the binary is executable
	err = os.Chmod(extractedToolPath, 0755)

	return
}

// getToolBinaryVersion returns the version of the tool binary.
func getToolBinaryVersion(
	toolPath string, versionArgs []string,
) (version *semver.Version, err error) {
	versionRegex := `(\d+\.\d+\.\d+)`

	r, err := regexp.Compile(versionRegex)
	if err != nil {
		return
	}

	out, err := exec.Command(toolPath, versionArgs...).Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			err = fmt.Errorf("❌ Could not determine installed version: %s (%w)",
				strings.TrimSpace(string(exitError.Stderr)), err,
			)
		}
		return
	}

	rawVersion := strings.TrimSpace(string(out))

	match := r.FindStringSubmatch(rawVersion)
	if len(match) < 2 {
		err = fmt.Errorf("could not find version in output: %s", rawVersion)
		return
	}

	version, err = semver.NewVersion(match[1])
	return
}

func isToolInstalled(toolName string) (installed bool, err error) {
	installedToolPath, err := which(toolName)
	if err != nil {
		return
	}
	if installedToolPath != "" {
		installed = true
	}
	return
}

func prependToolName(tool api.Tool, allTools []api.Tool, message ...string) string {
	if len(allTools) == 1 {
		return strings.Join(message, " ")
	}

	longestToolNameLength := 0
	for _, tool := range allTools {
		if len(tool.Name) > longestToolNameLength {
			longestToolNameLength = len(tool.Name)
		}
	}

	return "[" + tool.Name +
		strings.Repeat(" ", longestToolNameLength-len(tool.Name)) +
		"] " +
		strings.Join(message, " ")
}

func stripVersionsFromArgs(args []string) []string {
	var strippedArgs []string
	for _, arg := range args {
		if strings.Contains(arg, "@") {
			strippedArgs = append(strippedArgs, strings.SplitN(arg, "@", 2)[0])
		} else {
			strippedArgs = append(strippedArgs, arg)
		}
	}
	return strippedArgs
}

func which(toolName string) (path string, err error) {
	which := exec.Command("which", toolName)
	out, err := which.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		}
		return
	}
	path = strings.TrimSpace(string(out))
	return
}

func wrapInQuotesIfContainsSpace(s string) string {
	if strings.Contains(s, " ") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}
