package cmd_test

import (
	"bytes"
	"testing"

	"github.com/toolctl/toolctl/internal/cmd"
)

func TestRootCmd(t *testing.T) {
	usage := `Usage:
  toolctl [flags]
  toolctl [command]

Examples:
  # Get information about a tool
  toolctl info kubectl

  # Install a tool
  toolctl install minikube

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  info        Information about one or more tools
  install     Install one or more tools
  list        List the tools
  upgrade     Upgrade one or more tools
  version     Display the version of toolctl

Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
  -h, --help            help for toolctl
      --version         display the version of toolctl

Use "toolctl [command] --help" for more information about a command.
`

	tests := []test{
		{
			name:    "no cli args",
			cliArgs: []string{},
			wantOut: `toolctl controls your tools

` + usage,
		},
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `toolctl controls your tools

` + usage,
		},
		{
			name:    "config flag without value",
			cliArgs: []string{"--config"},
			wantErr: true,
			wantOut: `Error: flag needs an argument: --config
` + usage + `
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, nil)
			command.SetArgs(tt.cliArgs)

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
