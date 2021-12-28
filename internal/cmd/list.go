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
	allFlag      bool
	markdownFlag bool
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
	listCmd.Flags().BoolVarP(
		&allFlag, "all", "a", false,
		"list all supported tools, including those not installed",
	)

	// Hidden flags
	listCmd.Flags().BoolVar(
		&markdownFlag, "markdown", false,
		"output in markdown format",
	)
	err := listCmd.Flags().MarkHidden("markdown")
	if err != nil {
		panic(err)
	}

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

		err = list(cmd, toolctlWriter, toolctlAPI)
		if err != nil {
			return
		}

		return
	}
}

func list(
	cmd *cobra.Command, toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI,
) (err error) {
	// Get the metadata that holds the list of supported tools
	meta, err := api.GetMeta(toolctlAPI)
	if err != nil {
		return
	}

	// Build the list of tools
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

		if len(toolNames) == 0 {
			fmt.Fprintln(toolctlWriter, "No tools installed")
			return
		}
	}

	if markdownFlag {
		var localFlag bool
		localFlag, err = cmd.Flags().GetBool("local")
		if err != nil {
			return
		}

		if !localFlag {
			return fmt.Errorf("--markdown also requires --local")
		}

		return printMarkdown(toolctlWriter, toolctlAPI, toolNames)
	}

	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return printToPipe(toolctlWriter, toolNames)
	}

	return printToTerminal(toolctlWriter, toolNames)
}

func printToPipe(toolctlWriter io.Writer, toolNames []string) (err error) {
	for _, toolName := range toolNames {
		fmt.Fprintln(toolctlWriter, toolName)
	}
	return
}

func printToTerminal(toolctlWriter io.Writer, toolNames []string) (err error) {
	maxToolNameLen := 0
	for _, toolName := range toolNames {
		if len(toolName) > maxToolNameLen {
			maxToolNameLen = len(toolName)
		}
	}

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

func printMarkdown(
	toolctlWriter io.Writer, toolctlAPI api.ToolctlAPI, toolNames []string,
) (err error) {
	for _, toolName := range toolNames {
		var toolMeta api.ToolMeta
		toolMeta, err = api.GetToolMeta(toolctlAPI, api.Tool{Name: toolName})
		if err != nil {
			return
		}

		fmt.Fprintf(
			toolctlWriter, "- [%s](%s): %s\n",
			toolName, toolMeta.Homepage, toolMeta.Description,
		)
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
