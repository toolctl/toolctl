package api_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/spf13/afero"
	"github.com/toolctl/toolctl/internal/api"
)

const localAPIBasePath = "/toolctl/tools/v0"

type apiContents []apiFile

type apiFile struct {
	Path     string
	Contents string
}

func setupTest(apiLocation api.Location, apiContents apiContents) (
	toolctlAPI api.ToolctlAPI, apiServer *httptest.Server, err error,
) {
	localAPIFS := afero.NewMemMapFs()

	for _, f := range apiContents {
		err = afero.WriteFile(localAPIFS, f.Path, []byte(f.Contents), 0644)
		if err != nil {
			return
		}
	}

	if apiLocation == api.Remote {
		apiFileServer := http.FileServer(afero.NewHttpFs(localAPIFS).Dir(localAPIBasePath))
		apiServer = httptest.NewServer(apiFileServer)

		toolctlAPI, err = api.NewRemoteAPI(localAPIFS, apiServer.URL)
		if err != nil {
			return
		}

		return
	}

	if apiLocation == api.Local {
		toolctlAPI, err = api.NewLocalAPI(localAPIFS, localAPIBasePath)
		if err != nil {
			return
		}
	}

	return
}
