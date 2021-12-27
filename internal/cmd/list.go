package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

var (
	allFlag bool
)

func newListCmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the tools",
		Example: `  # List all installed tools
  toolctl list
  toolctl ls

  # List all supported tools, including those not installed
  toolctl list --all
  toolctl ls -a`,
		Args: cobra.NoArgs,
		RunE: newRunList(toolctlWriter, localAPIFS),
	}

	// Flags
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "list all supported tools, including those not installed")

	return listCmd
}

func newRunList(
	toolctlWriter io.Writer, localAPIFS afero.Fs,
) func(*cobra.Command, []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Remote)
		if err != nil {
			return err
		}

		err = list(toolctlWriter, toolctlAPI)
		if err != nil {
			return
		}

		return
	}
}

func list(toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI) (err error) {
	// Get the metadata that holds the list of supported tools
	meta, err := api.GetMeta(toolctlAPI)
	if err != nil {
		return
	}

	var toolNames []string
	if allFlag {
		toolNames = meta.Tools
	} else {
		for _, toolName := range meta.Tools {
			var installed bool
			installed, err = isToolInstalled(toolName)
			if err != nil {
				return
			}
			if installed {
				toolNames = append(toolNames, toolName)
			}
		}
	}

	// Print the list of tools
	if len(toolNames) == 0 {
		fmt.Fprintln(toolctlWriter, "No tools installed")
		return
	}

	maxToolNameLen := 0
	for _, toolName := range toolNames {
		if len(toolName) > maxToolNameLen {
			maxToolNameLen = len(toolName)
		}
	}

	// Set the output mode
	var outputMode string
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		outputMode = "pipe"
	} else {
		outputMode = "terminal"
	}

	// If the output is a terminal, apply some fancy formatting
	if outputMode == "terminal" {
		var terminalWidth int
		terminalWidth, err = getTerminalWidth()
		if err != nil {
			return
		}

		toolsPerLine := terminalWidth / (maxToolNameLen + 3)

		for i, toolName := range toolNames {
			fmt.Fprintf(
				toolctlWriter, "%s%s   ",
				toolName, strings.Repeat(" ", maxToolNameLen-len(toolName)),
			)
			if (i+1)%toolsPerLine == 0 {
				fmt.Fprintln(toolctlWriter)
			}
		}
		fmt.Fprintln(toolctlWriter)
		return
	}

	// Print the list of installed tools, one per line
	for _, toolName := range toolNames {
		fmt.Fprintln(toolctlWriter, toolName)
	}

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

func getTerminalWidth() (width int, err error) {
	// Get the terminal size
	var cmdOut []byte
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	cmdOut, err = cmd.Output()
	if err != nil {
		return
	}

	// Split the output into the width and height
	heightWidth := strings.Split(strings.TrimSpace(string(cmdOut)), " ")
	if len(heightWidth) != 2 {
		err = fmt.Errorf("invalid output from stty size")
		return
	}

	width, err = strconv.Atoi(heightWidth[1])
	if err != nil {
		return
	}

	return
}
