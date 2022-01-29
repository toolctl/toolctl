package utils_test

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/toolctl/toolctl/internal/utils"
)

func TestState(t *testing.T) {
	localFS := afero.NewMemMapFs()

	state, err := utils.NewState(localFS)
	if err != nil {
		t.Errorf("NewState() error = %v, wantErr <nil>", err)
		return
	}

	upgradeLastSuccess := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	state.Upgrade.LastSuccess = upgradeLastSuccess
	if err := state.Write(localFS); err != nil {
		t.Errorf("state.Write() error = %v, wantErr <nil>", err)
		return
	}

	state, err = utils.NewState(localFS)
	if err != nil {
		t.Errorf("NewState() error = %v, wantErr <nil>", err)
		return
	}

	if got, want := state.Upgrade.LastSuccess, upgradeLastSuccess; got != want {
		t.Errorf("state.Upgrade.LastSuccess = %s, want %s", got, want)
	}
}
