package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/api"
)

func newSyncCmd(localAPIFS afero.Fs) *cobra.Command {
	discoverCmd := &cobra.Command{
		Use:   "sync [flags]",
		Short: "Sync the list of supported tools",
		Args:  cobra.NoArgs,
		RunE:  newRunSync(localAPIFS),
	}

	return discoverCmd
}

func newRunSync(
	localAPIFS afero.Fs,
) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, _ []string) (err error) {
		// Needs to run with the local API because we need write access
		toolctlAPI, err := api.New(localAPIFS, cmd, api.Local)
		if err != nil {
			return
		}

		// Detect all tool directories
		tools := []string{}
		matches, err := afero.ReadDir(localAPIFS, toolctlAPI.LocalAPIBasePath())
		if err != nil {
			return
		}
		for _, match := range matches {
			if match.IsDir() {
				tools = append(tools, match.Name())
			}
		}

		// Save the metadata
		err = api.SaveMeta(toolctlAPI, api.Meta{Tools: tools})
		return
	}
}
