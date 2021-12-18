package api

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/toolctl/toolctl/internal/utils"
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
	GetContents(path string) (found bool, contents []byte, err error)
	GetLocalAPIFS() afero.Fs
	GetLocation() Location

	SaveContents(path string, data []byte) error
}

// New returns a new API instance, based on the specified command line flags
// and the default location.
func New(localAPIFS afero.Fs, cmd *cobra.Command, defaultLocation Location) (ToolctlAPI, error) {
	localFlag, err := cmd.Flags().GetBool("local")
	if err != nil {
		return nil, err
	}

	if localFlag || defaultLocation == Local {
		localAPIBasePath, err := utils.RequireConfigString("LocalAPIBasePath")
		if err != nil {
			return nil, err
		}
		return NewLocalAPI(localAPIFS, localAPIBasePath)
	}

	remoteAPIBaseURL, err := utils.RequireConfigString("RemoteAPIBaseURL")
	if err != nil {
		return nil, err
	}
	return NewRemoteAPI(localAPIFS, remoteAPIBaseURL)

}
