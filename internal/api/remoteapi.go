package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/afero"
)

type remoteAPI struct {
	BaseURL    *url.URL
	LocalAPIFS afero.Fs
}

func (a remoteAPI) GetContents(path string) (found bool, contents []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(a.BaseURL.ResolveReference(&url.URL{Path: path}).String())
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

// GetLocalFS returns the underlying local filesystem.
func (a remoteAPI) GetLocalAPIFS() (fs afero.Fs) {
	return a.LocalAPIFS
}

func (a remoteAPI) GetLocation() Location {
	return Remote
}

// The remote API is currently read-only.
func (a remoteAPI) SaveContents(path string, contents []byte) (err error) {
	return fmt.Errorf("not implemented")
}

func NewRemoteAPI(localFS afero.Fs, remoteAPIBaseURL string) (ToolctlAPI, error) {
	var baseURL *url.URL
	baseURL, err := url.Parse(remoteAPIBaseURL)
	if err != nil {
		return nil, err
	}

	return &remoteAPI{
		LocalAPIFS: localFS,
		BaseURL:    baseURL,
	}, nil
}
