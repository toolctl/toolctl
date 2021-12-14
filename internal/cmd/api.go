// Package ops contains all the operations-related commands.
package cmd

import (
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newAPICmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	apiCmd := &cobra.Command{
		Use:    "api",
		Short:  "Commands for managing the toolctl API",
		Hidden: true,
	}

	apiCmd.AddCommand(newDiscoverCmd(toolctlWriter, localAPIFS))

	return apiCmd
}
