// Package api contains everything related to the toolctl API.
package api

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/sysutil"
)

// Location represents the location of the API, currently remote or local.
type Location uint32

const (
	// Local represents the local API location.
	Local Location = iota
	// Remote represents the remote API location.
	Remote
)

// ToolctlAPI defines the interface that all toolctl APIs need to implement.
type ToolctlAPI interface {
	LocalAPIBasePath() string
	LocalAPIFS() afero.Fs
	Location() Location

	GetContents(path string) (found bool, contents []byte, err error)
	SaveContents(path string, data []byte) error
}

// New returns a new API instance, based on the specified command line flags
// and the default location.
func New(
	localAPIFS afero.Fs, cmd *cobra.Command, defaultLocation Location,
) (toolctlAPI ToolctlAPI, err error) {
	var localFlag bool
	localFlag, err = cmd.Flags().GetBool("local")
	if err != nil {
		return
	}

	if localFlag || defaultLocation == Local {
		var localAPIBasePath string
		localAPIBasePath, err = sysutil.RequireConfigString("LocalAPIBasePath")
		if err != nil {
			return
		}
		return NewLocalAPI(localAPIFS, localAPIBasePath)
	}

	var remoteAPIBaseURL string
	remoteAPIBaseURL, err = sysutil.RequireConfigString("RemoteAPIBaseURL")
	if err != nil {
		return
	}
	return NewRemoteAPI(localAPIFS, remoteAPIBaseURL)
}
