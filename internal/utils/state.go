package utils

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/spf13/afero"
	"github.com/twpayne/go-xdg/v6"
)

// A State records state.
type State struct {
	filename string
	Upgrade  struct {
		LastSuccess time.Time
	}
}

// NewState returns a state persisted to localFS. If there is no existing state,
// then an empty state is returned.
func NewState(localFS afero.Fs) (state *State, err error) {
	// Follow the XDG Base Directory Specification to determine the state's
	// location
	var bds *xdg.BaseDirectorySpecification
	bds, err = xdg.NewBaseDirectorySpecification()
	if err != nil {
		return
	}

	// Create an empty state at the default location
	state = &State{
		filename: path.Join(bds.CacheHome, "toolctl", "state.json"),
	}

	// Open the state file
	var file afero.File
	file, err = localFS.Open(state.filename)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// If the state file does not exist then return the empty state, without
		// error
		err = nil
		return
	case err != nil:
		return
	}

	// Unmarshal the state
	var data []byte
	data, err = io.ReadAll(file)
	_ = file.Close()
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		return
	}

	return
}

// Write writes s to localFS.
func (s *State) Write(localFS afero.Fs) (err error) {
	// Create parent directories if needed
	err = localFS.MkdirAll(path.Dir(s.filename), 0o777)
	if err != nil {
		return
	}

	// Create or replace the state file
	var file afero.File
	file, err = localFS.OpenFile(s.filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		return
	}
	err = json.NewEncoder(file).Encode(s)
	_ = file.Close()
	return err
}
