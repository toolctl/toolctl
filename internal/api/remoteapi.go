package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/afero"
)

type remoteAPI struct {
	baseURL    *url.URL
	localAPIFS afero.Fs
}

func (a remoteAPI) GetContents(path string) (found bool, contents []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(a.baseURL.ResolveReference(&url.URL{Path: path}).String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	contents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	found = true

	return
}

func (a remoteAPI) LocalAPIBasePath() string {
	return ""
}

func (a remoteAPI) LocalAPIFS() (fs afero.Fs) {
	return a.localAPIFS
}

func (a remoteAPI) Location() Location {
	return Remote
}

// SaveContents is currently not supported by the remote API.
func (a remoteAPI) SaveContents(path string, contents []byte) (err error) {
	return fmt.Errorf("not implemented")
}

// NewRemoteAPI returns a new remote API instance.
func NewRemoteAPI(localFS afero.Fs, remoteAPIBaseURL string) (ToolctlAPI, error) {
	var baseURL *url.URL
	baseURL, err := url.Parse(remoteAPIBaseURL)
	if err != nil {
		return nil, err
	}

	return &remoteAPI{
		localAPIFS: localFS,
		baseURL:    baseURL,
	}, nil
}
