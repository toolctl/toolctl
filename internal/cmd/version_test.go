package cmd_test

import (
	"bytes"
	"testing"

	"github.com/toolctl/toolctl/internal/cmd"
)

func TestVersionCmd(t *testing.T) {
	usage := `Usage:
  toolctl version [flags]

Flags:
  -h, --help    help for version
      --short   display only the version number

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
`

	tests := []test{
		{
			name:    "no cli args",
			cliArgs: []string{},
			wantOut: `{"GitVersion":"v0.0.0-dev","GitCommit":"da39a3ee5e6b4b0d3255bfef95601890afd80709","BuildDate":"0000-00-00T00:00:00Z"}
`,
		},
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `Display the version of toolctl

` + usage,
		},
		{
			name:    "short flag",
			cliArgs: []string{"--short"},
			wantOut: `v0.0.0-dev
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, nil)
			command.SetArgs(append([]string{"version"}, tt.cliArgs...))

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: Execute() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			checkWantOut(t, tt, buf)
		})
	}
}
