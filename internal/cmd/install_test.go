package cmd_test

import (
	"runtime"
	"testing"
)

func TestInstallCmd(t *testing.T) {
	usage := `Usage:
  toolctl install TOOL[@VERSION]... [flags]

Examples:
  # Install the latest version of a tool
  toolctl install minikube

  # Install a specified version of a tool
  toolctl install kubectl@1.20.13

  # Install multiple tools
  toolctl install gh k9s

Flags:
  -h, --help   help for install

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
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `👷 Installing v0.1.1 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool as .tar.gz",
			cliArgs: []string{"toolctl-test-tool-tar-gz"},
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool-tar-gz",
					version: "0.1.0",
					tarGz:   true,
				},
			},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool as .tar.gz in subdir",
			cliArgs: []string{"toolctl-test-tool-tar-gz"},
			supportedTools: []supportedTool{
				{
					name:        "toolctl-test-tool-tar-gz",
					version:     "0.1.0",
					tarGz:       true,
					tarGzSubdir: runtime.GOOS + "-" + runtime.GOARCH,
				},
			},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool as .tar.gz in random subdir with dashed platform suffix",
			cliArgs: []string{"toolctl-test-tool-tar-gz"},
			supportedTools: []supportedTool{
				{
					name:        "toolctl-test-tool-tar-gz",
					version:     "0.1.0",
					tarGz:       true,
					tarGzSubdir: "out",
					tarGzBinaryName: "toolctl-test-tool-tar-gz" + "-" +
						runtime.GOOS + "-" + runtime.GOARCH,
				},
			},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool as .tar.gz in random subdir with underscored platform suffix",
			cliArgs: []string{"toolctl-test-tool-tar-gz"},
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool-tar-gz",
					version: "0.1.0",
					tarGz:   true,
					tarGzSubdir: "toolctl-test-tool-tar-gz" + "_" +
						runtime.GOOS + "_" + runtime.GOARCH,
					tarGzBinaryName: "toolctl-test-tool-tar-gz" + "_" +
						runtime.GOOS + "_" + runtime.GOARCH,
				},
			},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool with supported version",
			cliArgs: []string{"toolctl-test-tool@0.1.0"},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "multiple supported tools",
			cliArgs: []string{"toolctl-test-tool@0.1.0", "toolctl-test-tool@0.1.1"},
			wantOut: `[toolctl-test-tool] 👷 Installing v0.1.0 ...
[toolctl-test-tool] 🎉 Successfully installed
[toolctl-test-tool] 🤷 v0.1.0 is already installed
[toolctl-test-tool] 💁 For more details run: toolctl info toolctl-test-tool
`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "supported tool with unsupported version",
			cliArgs: []string{"toolctl-test-tool@1.0.0"},
			wantErr: true,
			wantOut: `👷 Installing v1.0.0 ...
Error: toolctl-test-tool v1.0.0 could not be found
`,
		},
		// -------------------------------------------------------------------------
		{
			name:               "install dir could not be found",
			cliArgs:            []string{"toolctl-test-tool"},
			installDirNotFound: true,
			wantErr:            true,
			wantOutRegex: `^Error: install directory .+toolctl-test-install-\d+-nonexistent does not exist
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:                "install dir not in path",
			cliArgs:             []string{"toolctl-test-tool"},
			installDirNotInPath: true,
			wantOutRegex: `^🚨 .+toolctl-test-install-\d+ is not in \$PATH
👷 Installing v0.1.1 ...
🎉 Successfully installed
$`,
		},
		// -------------------------------------------------------------------------
		{
			name:                  "install dir not writable",
			cliArgs:               []string{"a-tool", "another-tool@0.1.2"},
			installDirNotWritable: true,
			wantErr:               true,
			wantOutRegex: `^Error: .+toolctl-test-install-\d+ is not writable by user .+, try running:
  sudo toolctl install a-tool another-tool@0.1.2
$`,
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
			wantOut: `🤷 v0.1.1 (the latest version) is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
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
			wantOut: `🤷 v0.1.0 is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
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
			wantOut: `🤷 Unknown version is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
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
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool with version mismatch and supported tool",
			supportedTools: []supportedTool{
				{
					name:          "toolctl-test-tool-version-mismatch",
					version:       "0.1.2",
					binaryVersion: "0.1.1",
					tarGz:         true,
				},
			},
			cliArgs: []string{
				"toolctl-test-tool-version-mismatch", "toolctl-other-test-tool",
			},
			wantErr: true,
			wantOut: `[toolctl-test-tool-version-mismatch] 👷 Installing v0.1.2 ...
Error: installation failed: expected v0.1.2, but installed binary reported v0.1.1
`,
		},
	}

	runInstallUpgradeTests(t, tests, "install")
}
