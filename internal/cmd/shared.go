package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

// ArgToTool converts a command line argument to a tool.
func ArgToTool(arg string, versionAllowed bool) (tool api.Tool, err error) {
	splitArg := strings.SplitN(arg, "@", 2)
	tool = api.Tool{
		Name: splitArg[0],
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
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
func ArgsToTools(args []string, versionAllowed bool) ([]api.Tool, error) {
	var tools []api.Tool

	for _, arg := range args {
		tool, err := ArgToTool(arg, versionAllowed)
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

func checkArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			err = fmt.Errorf("no tool specified")
		}
		return
	}
}

// getLatestVersion returns the version of an installed tool.
func getInstalledVersion(
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

	match := r.FindStringSubmatch(string(out))
	if len(match) < 2 {
		err = fmt.Errorf("could not find version in output: %s", string(out))
		return
	}

	version, err = semver.NewVersion(match[1])
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

// toolToArg converts a tool to a command line argument.
func toolToArg(tool api.Tool) (arg string) {
	arg = tool.Name
	if tool.Version != "" {
		arg += "@" + tool.Version
	}
	return
}

// toolsToArgs converts a list of tools to a list of command line arguments.
func toolsToArgs(tools []api.Tool) (args string) {
	argsSlice := make([]string, len(tools))
	for i, tool := range tools {
		argsSlice[i] = toolToArg(tool)
	}
	args = strings.Join(argsSlice, " ")
	return
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
