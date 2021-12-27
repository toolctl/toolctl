package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
)

func TestInfoCmd(t *testing.T) {
	tests := []test{
		{
			name:    "--help flag",
			cliArgs: []string{"--help"},
			wantOut: `Get information about tools

Usage:
  toolctl info [TOOL...] [flags]

Examples:
  # Get information about installed tools
  toolctl info

  # Get information about a specific tool
  toolctl info kubectl

  # Get information about multiple tools
  toolctl info gh k9s

Flags:
  -h, --help   help for info

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "no args, no supported tools installed",
			cliArgs: []string{},
			wantErr: true,
			wantOut: `Error: no supported tools installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "no args, supported tool installed",
			cliArgs: []string{"toolctl-test-tool"},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			wantOutRegex: `^✨ toolctl-test-tool v0.1.1: toolctl test tool
🔄 toolctl-test-tool v0.1.0 is installed at .+
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			wantOut: `✨ toolctl-test-tool v0.1.1: toolctl test tool
🏠 https://toolctl.io/
❌ Not installed
`,
		},
		// -------------------------------------------------------------------------
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
		// -------------------------------------------------------------------------
		{
			name:    "multiple supported tools, one of them with version",
			cliArgs: []string{"toolctl-test-tool", "toolctl-test-tool@2.0.0"},
			wantErr: true,
			wantOut: `Error: please don't specify a tool version, try this instead:
  toolctl info toolctl-test-tool toolctl-test-tool
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool, latest version already installed",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
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
		// -------------------------------------------------------------------------
		{
			name: "supported tool, other version already installed",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
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
		// -------------------------------------------------------------------------
		{
			name: "supported tool, version could not be determined",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
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
		// -------------------------------------------------------------------------
		{
			name: "supported tool symlinked",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.1"
`,
				},
			},
			preinstalledToolIsSymlinked: true,
			cliArgs:                     []string{"toolctl-test-tool"},
			wantOutRegex: `✨ toolctl-test-tool v0.1.1: toolctl test tool
✅ toolctl-test-tool v0.1.1 is installed at .+
🔗 Symlinked from .+/symlinked-toolctl-test-tool
$`,
		},
		// -------------------------------------------------------------------------
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
		toolctlAPI, apiServer, downloadServer, err := setupRemoteAPI(tt.supportedTools)
		if err != nil {
			t.Fatal(err)
		}

		var preinstallTempDir string
		if !cmp.Equal(tt.preinstalledTools, []preinstalledTool{}) {
			preinstallTempDir = setupPreinstallTempDir(
				t, tt, toolctlAPI, originalPathEnv,
			)
		}

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, toolctlAPI.LocalAPIFS())
			command.SetArgs(append([]string{"info"}, tt.cliArgs...))
			viper.Set("RemoteAPIBaseURL", apiServer.URL)

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}

			checkWantOut(t, tt, buf)
		})

		os.Setenv("PATH", originalPathEnv)

		if !cmp.Equal(tt.preinstalledTools, []preinstalledTool{}) {
			err = os.RemoveAll(preinstallTempDir)
			if err != nil {
				t.Fatal(err)
			}
		}

		apiServer.Close()
		downloadServer.Close()
	}
}
