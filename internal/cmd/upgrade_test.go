package cmd_test

import (
	"testing"
)

func TestUpgradeCmd(t *testing.T) {
	usage := `Usage:
  toolctl upgrade [TOOL...] [flags]

Examples:
  # Upgrade all tools
  toolctl upgrade

  # Upgrade a specific tool
  toolctl upgrade minikube

  # Upgrade multiple tools
  toolctl upgrade gh k9s

Flags:
  -h, --help   help for upgrade

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
`

	tests := []test{
		{
			name:    "--help flag",
			cliArgs: []string{"--help"},
			wantOut: "Upgrade tools\n\n" + usage,
		},
		// -------------------------------------------------------------------------
		{
			name: "no args",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			wantOut: `👷 Upgrading from v0.1.0 to v0.1.1 ...
👷 Removing v0.1.0 ...
👷 Installing v0.1.1 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
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
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
				{
					name:    "toolctl-other-test-tool",
					version: "0.2.0",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
				{
					name: "toolctl-other-test-tool",
					fileContents: `#!/bin/sh
echo "v0.2.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool", "toolctl-other-test-tool"},
			wantOut: `[toolctl-test-tool      ] 👷 Upgrading from v0.1.0 to v0.1.1 ...
[toolctl-test-tool      ] 👷 Removing v0.1.0 ...
[toolctl-test-tool      ] 👷 Installing v0.1.1 ...
[toolctl-test-tool      ] 🎉 Successfully installed
[toolctl-other-test-tool] ✅ Already up to date (v0.2.0)
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool, latest version already installed",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.1"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `✅ Already up to date (v0.1.1)
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool, not installed",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-test-tool is not installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool, symlinked",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
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
			wantOutRegex: `^🚫 Skipping: .+ is symlinked from .+
$`,
		},
		// -------------------------------------------------------------------------
		{
			name: "install dir not writable, no tools specified",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs:               []string{},
			installDirNotWritable: true,
			wantErr:               true,
			wantOutRegex: `^Error: .+toolctl-test-install-\d+ is not writable by user .+, try running:
  sudo toolctl upgrade
$`,
		},
		// -------------------------------------------------------------------------
		{
			name: "install dir not writable, tool specified",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs:               []string{"toolctl-test-tool"},
			installDirNotWritable: true,
			wantErr:               true,
			wantOutRegex: `^Error: .+toolctl-test-install-\d+ is not writable by user .+, try running:
  sudo toolctl upgrade toolctl-test-tool
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:                       "supported tool, installed not in install dir",
			installDirNotPreinstallDir: true,
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.1",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.1.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantOutRegex: `^🚫 Skipping: toolctl-test-tool is installed in .+, not in .+
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
			name: "supported tool, installed version newer than latest",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
			},
			preinstalledTools: []preinstalledTool{
				{
					name: "toolctl-test-tool",
					fileContents: `#!/bin/sh
echo "v0.2.0"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `🚫 Skipping: toolctl-test-tool is already at v0.2.0, but the latest version is v0.1.0
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
