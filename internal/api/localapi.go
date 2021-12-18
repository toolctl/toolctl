package api

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
)

type localAPI struct {
	BasePath   string
	LocalAPIFS afero.Fs
}

func (a localAPI) GetContents(relativePath string) (found bool, contents []byte, err error) {
	absolutePath := path.Join(a.BasePath, relativePath)

	contents, err = afero.ReadFile(a.LocalAPIFS, absolutePath)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
		return
	}
	found = true

	return
}

func (a localAPI) GetLocalAPIFS() afero.Fs {
	return a.LocalAPIFS
}

func (a localAPI) GetLocation() Location {
	return Local
}

// Save writes the give contents to a file at the given path.
func (a localAPI) SaveContents(relativePath string, contents []byte) (err error) {
	absolutePath := path.Join(a.BasePath, relativePath)
	err = a.LocalAPIFS.MkdirAll(filepath.Dir(absolutePath), 0755)
	if err != nil {
		return
	}
	err = afero.WriteFile(a.LocalAPIFS, absolutePath, contents, 0644)
	if err != nil {
		return
	}
	return
}

// NewLocalAPI returns a new local API instance.
func NewLocalAPI(localAPIFS afero.Fs, basePath string) (ToolctlAPI, error) {
	return &localAPI{
		BasePath:   basePath,
		LocalAPIFS: localAPIFS,
	}, nil
}
