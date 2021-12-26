package api

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
)

type localAPI struct {
	basePath   string
	localAPIFS afero.Fs
}

func (a localAPI) GetContents(relativePath string) (found bool, contents []byte, err error) {
	absolutePath := path.Join(a.basePath, relativePath)

	contents, err = afero.ReadFile(a.localAPIFS, absolutePath)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
		return
	}
	found = true

	return
}

func (a localAPI) LocalAPIBasePath() string {
	return a.basePath
}

func (a localAPI) LocalAPIFS() afero.Fs {
	return a.localAPIFS
}

func (a localAPI) Location() Location {
	return Local
}

// Save writes the give contents to a file at the given path.
func (a localAPI) SaveContents(relativePath string, contents []byte) (err error) {
	absolutePath := path.Join(a.basePath, relativePath)
	err = a.localAPIFS.MkdirAll(filepath.Dir(absolutePath), 0755)
	if err != nil {
		return
	}
	err = afero.WriteFile(a.localAPIFS, absolutePath, contents, 0644)
	if err != nil {
		return
	}
	return
}

// NewLocalAPI returns a new local API instance.
func NewLocalAPI(localAPIFS afero.Fs, basePath string) (ToolctlAPI, error) {
	return &localAPI{
		basePath:   basePath,
		localAPIFS: localAPIFS,
	}, nil
}
