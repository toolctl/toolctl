package cmd_test

import (
	"testing"
)

func TestUpgradeCmd(t *testing.T) {
	usage := `Usage:
  toolctl upgrade TOOL... [flags]

Examples:
  # Upgrade a tool
  toolctl upgrade minikube

  # Upgrade multiple tools
  toolctl upgrade kustomize k9s

Flags:
  -h, --help   help for upgrade

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)

`

	tests := []test{
		{
			name:    "no cli args",
			cliArgs: []string{},
			wantErr: true,
			wantOut: `Error: no tool specified
` + usage,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `👷 Upgrading from v0.1.0 to v0.1.1 ...
👷 Removing v0.1.0 ...
👷 Installing v0.1.1 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool with version",
			cliArgs: []string{"toolctl-test-tool@0.1.0"},
			wantErr: true,
			wantOut: `Error: please don't specify a tool version, try this instead:
  toolctl upgrade toolctl-test-tool
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "multiple supported tools",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool", "toolctl-test-tool"},
			wantOut: `[toolctl-test-tool] 👷 Upgrading from v0.1.0 to v0.1.1 ...
[toolctl-test-tool] 👷 Removing v0.1.0 ...
[toolctl-test-tool] 👷 Installing v0.1.1 ...
[toolctl-test-tool] 🎉 Successfully installed
[toolctl-test-tool] ✅ already up-to-date
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
			wantOut: `✅ already up-to-date
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool, not installed",
			cliArgs: []string{"toolctl-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-test-tool is not installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool, symlinked",
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			preinstalledToolIsSymlinked: true,
			cliArgs:                     []string{"toolctl-test-tool"},
			wantErr:                     true,
			wantOutRegex: `Error: aborting: .+ is symlinked from .+
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:                  "install dir not writable",
			cliArgs:               []string{"a-tool", "another-tool"},
			installDirNotWritable: true,
			wantErr:               true,
			wantOutRegex: `^Error: .+toolctl-test-install-\d+ is not writable by user .+, try running:
  sudo toolctl upgrade a-tool another-tool
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:                       "supported tool, installed not in install dir",
			installDirNotPreinstallDir: true,
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			preinstalledToolIsSymlinked: true,
			cliArgs:                     []string{"toolctl-test-tool"},
			wantErr:                     true,
			wantOutRegex: `Error: aborting: toolctl-test-tool is currently installed in .+, not in .+
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
		// -------------------------------------------------------------------------
		{
			name:    "unsupported tool with version",
			cliArgs: []string{"toolctl-unsupported-test-tool@1.0.0"},
			wantErr: true,
			wantOut: `Error: please don't specify a tool version, try this instead:
  toolctl upgrade toolctl-unsupported-test-tool
`,
		},
	}

	runInstallUpgradeTests(t, tests, "upgrade")
}
