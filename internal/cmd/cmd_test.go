package cmd_test

import (
	"testing"

	"github.com/toolctl/toolctl/internal/cmd"
)

func TestExecute(t *testing.T) {
	tests := []test{
		{
			name: "should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			cmd.Execute()
		})
	}
}
