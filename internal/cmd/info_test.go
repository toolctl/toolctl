package cmd_test

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
)

func TestInfoCmd(t *testing.T) {
	tests := []test{
		{
			name:    "no cli args",
			cliArgs: []string{},
			wantErr: true,
			wantOut: `Error: no tool specified
Usage:
  toolctl info TOOL... [flags]

Examples:
  # Get information about a tool
  toolctl info kubectl

  # Get information about multiple tools
  toolctl info k9s kustomize

Flags:
  -h, --help   help for info

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)

`,
		},
		{
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `✨ toolctl-test-tool v0.1.1: toolctl test tool
🏠 https://toolctl.io/
❌ Not installed
`,
		},
		{
			name:    "multiple supported tools",
			cliArgs: []string{"toolctl-test-tool", "toolctl-test-tool"},
			wantOut: `[toolctl-test-tool] ✨ toolctl-test-tool v0.1.1: toolctl test tool
[toolctl-test-tool] 🏠 https://toolctl.io/
[toolctl-test-tool] ❌ Not installed
[toolctl-test-tool] ✨ toolctl-test-tool v0.1.1: toolctl test tool
[toolctl-test-tool] 🏠 https://toolctl.io/
[toolctl-test-tool] ❌ Not installed
`,
		},
		{
			name:    "multiple supported tools, one of them with version",
			cliArgs: []string{"toolctl-test-tool", "toolctl-test-tool@2.0.0"},
			wantErr: true,
			wantOut: `Error: please don't specify a tool version, try this instead:
  toolctl info toolctl-test-tool toolctl-test-tool
`,
		},
		{
			name: "supported tool, latest version already installed",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/sh
echo "v0.1.1"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantErr: false,
			wantOutRegex: `✨ toolctl-test-tool v0.1.1: toolctl test tool
✅ toolctl-test-tool v0.1.1 is installed at .+
$`,
		},
		{
			name: "supported tool, other version already installed",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantErr: false,
			wantOutRegex: `✨ toolctl-test-tool v0.1.1: toolctl test tool
🔄 toolctl-test-tool v0.1.0 is installed at .+
$`,
		},
		{
			name: "supported tool, version could not be determined",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/sh
echo "version flag not supported" >&2
exit 1
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantErr: false,
			wantOut: `✨ toolctl-test-tool v0.1.1: toolctl test tool
❌ Could not determine installed version: version flag not supported (exit status 1)
`,
		},
		{
			name: "supported tool symlinked",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/sh
echo "v0.1.1"
`,
				},
			},
			preinstalledToolIsSymlinked: true,
			cliArgs:                     []string{"toolctl-test-tool"},
			wantErr:                     false,
			wantOutRegex: `✨ toolctl-test-tool v0.1.1: toolctl test tool
✅ toolctl-test-tool v0.1.1 is installed at .+
🔗 Symlinked from .+/symlinked-toolctl-test-tool
$`,
		},
		{
			name:    "unsupported tool",
			cliArgs: []string{"toolctl-unsupported-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
		{
			name:    "tool unsupported on current platform",
			cliArgs: []string{"toolctl-test-tool-unsupported-on-current-platform"},
			wantErr: true,
			wantOut: `Error: toolctl-test-tool-unsupported-on-current-platform not supported on this platform
`,
		},
	}

	originalPathEnv := os.Getenv("PATH")

	for _, tt := range tests {
		toolctlAPI, apiServer, downloadServer, err := setupRemoteAPI()
		if err != nil {
			t.Fatal(err)
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
			command.SetArgs(append([]string{"info"}, tt.cliArgs...))
			viper.Set("RemoteAPIBaseURL", apiServer.URL)

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantOut == "" && tt.wantOutRegex == "" {
				t.Fatalf("Either wantOut or wantOutRegex must be set")
			}
			if tt.wantOut != "" {
				if diff := cmp.Diff(tt.wantOut, buf.String()); diff != "" {
					t.Errorf("Output mismatch (-want +got):\n%s", diff)
				}
			} else if tt.wantOutRegex != "" {
				matched, err := regexp.Match(tt.wantOutRegex, buf.Bytes())
				if err != nil {
					t.Errorf("Error compiling regex: %v", err)
				}
				if !matched {
					t.Errorf("Error matching regex: %v, output: %s", tt.wantOutRegex, buf.String())
				}
			}
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
