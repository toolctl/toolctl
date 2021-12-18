package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
)

func TestListCmd(t *testing.T) {
	usage := `Usage:
  toolctl list [flags]

Aliases:
  list, ls

Examples:
  # List all installed tools
  toolctl list
  toolctl ls

  # List all supported tools, including those not installed
  toolctl list --all
  toolctl ls -a

Flags:
  -a, --all    list all supported tools, including those not installed
  -h, --help   help for list

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
`

	tests := []test{
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `List the tools

` + usage,
		},
		{
			name: "no tools installed",
			wantOut: `No tools installed
`,
		},
		{
			name: "toolctl-test-tool installed",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/bash
echo v0.1.0
`},
				{
					Name: "toolctl-another-test-tool",
					FileContents: `#!/bin/bash
echo v0.1.0
`},
			},

			wantOut: `toolctl-test-tool
toolctl-another-test-tool
`,
		},
	}

	originalPathEnv := os.Getenv("PATH")

	for _, tt := range tests {
		toolctlAPI, apiServer, downloadServer, err := setupRemoteAPI()
		if err != nil {
			t.Fatalf("%s: SetupRemoteAPI() failed: %v", tt.name, err)
		}

		var preinstalledTempInstallDir string
		if !cmp.Equal(tt.preinstalledTools, preinstalledTool{}) {
			preinstalledTempInstallDir, err = install(
				t, toolctlAPI, tt.preinstalledTools, tt.preinstalledToolIsSymlinked,
				originalPathEnv,
			)
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, toolctlAPI.GetLocalAPIFS())
			command.SetArgs(append([]string{"list"}, tt.cliArgs...))
			viper.Set("RemoteAPIBaseURL", apiServer.URL)

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: Execute() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			checkWantOut(t, tt, buf)
		})

		os.Setenv("PATH", originalPathEnv)

		if !cmp.Equal(tt.preinstalledTools, preinstalledTool{}) {
			err = os.RemoveAll(preinstalledTempInstallDir)
			if err != nil {
				t.Fatal(err)
			}
		}

		apiServer.Close()
		downloadServer.Close()
	}
}
