package cmd_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
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
  toolctl install kustomize k9s

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
		{
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `👷 Installing v0.1.1 ...
🎉 Successfully installed
`,
		},
		{
			name:    "supported tool as .tar.gz",
			cliArgs: []string{"toolctl-test-tool-tar-gz"},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		{
			name:    "supported tool as .tar.gz in subdir",
			cliArgs: []string{"toolctl-test-tool-tar-gz-subdir"},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		{
			name:    "supported tool with supported version",
			cliArgs: []string{"toolctl-test-tool@0.1.0"},
			wantOut: `👷 Installing v0.1.0 ...
🎉 Successfully installed
`,
		},
		{
			name:    "multiple supported tools",
			cliArgs: []string{"toolctl-test-tool@0.1.0", "toolctl-test-tool@0.1.1"},
			wantOut: `[toolctl-test-tool] 👷 Installing v0.1.0 ...
[toolctl-test-tool] 🎉 Successfully installed
[toolctl-test-tool] 🤷 v0.1.0 is already installed
[toolctl-test-tool] 💁 For more details run: toolctl info toolctl-test-tool
`,
		},
		{
			name:    "supported tool with unsupported version",
			cliArgs: []string{"toolctl-test-tool@1.0.0"},
			wantErr: true,
			wantOut: `👷 Installing v1.0.0 ...
Error: toolctl-test-tool v1.0.0 could not be found
`,
		},
		{
			name:               "install dir could not be found",
			cliArgs:            []string{"toolctl-test-tool"},
			installDirNotFound: true,
			wantErr:            true,
			wantOutRegex: `^Error: install directory .+toolctl-test-install-\d+-nonexistent does not exist
$`,
		},
		{
			name:                "install dir not in path",
			cliArgs:             []string{"toolctl-test-tool"},
			installDirNotInPath: true,
			wantOutRegex: `^🚨 .+toolctl-test-install-\d+ is not in \$PATH
👷 Installing v0.1.1 ...
🎉 Successfully installed
$`,
		},
		{
			name:                  "install dir not writable",
			cliArgs:               []string{"a-tool", "another-tool@0.1.2"},
			installDirNotWritable: true,
			wantErr:               true,
			wantOutRegex: `^Error: .+toolctl-test-install-\d+ is not writable by user .+, try running:
  sudo toolctl install a-tool another-tool@0.1.2
$`,
		},
		{
			name: "supported tool, latest valready installed",
			preinstalledTools: []preinstalledTool{
				{
					Name: "toolctl-test-tool",
					FileContents: `#!/bin/sh
echo "v0.1.1"
`,
				},
			},
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `🤷 v0.1.1 (the latest version) is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
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
			wantOut: `🤷 v0.1.0 is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
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
			wantOut: `🤷 Unknown version is already installed
💁 For more details run: toolctl info toolctl-test-tool
`,
		},
		{
			name:    "unsupported tool",
			cliArgs: []string{"toolctl-unsupported-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
		{
			name:    "unsupported tool with version",
			cliArgs: []string{"toolctl-unsupported-test-tool@1.0.0"},
			wantErr: true,
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
		{
			name: "supported tool with vmismatch and supported tool",
			cliArgs: []string{
				"toolctl-test-tool-version-mismatch", "toolctl-test-tool",
			},
			wantErr: true,
			wantOut: `[toolctl-test-tool-version-mismatch] 👷 Installing v0.1.0 ...
Error: installation failed: Expected v0.1.0, but installed v0.2.0
`,
		},
	}

	originalPathEnv := os.Getenv("PATH")

	for _, tt := range tests {
		toolctlAPI, apiServer, downloadServer, err := setupRemoteAPI()
		if err != nil {
			t.Fatal(err)
		}

		tempInstallDir, err := ioutil.TempDir("", "toolctl-test-install-*")
		if err != nil {
			t.Fatal(err)
		}

		if tt.installDirNotWritable {
			err = os.Chmod(tempInstallDir, 0500)
			if err != nil {
				t.Fatal(err)
			}
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
			command.SetArgs(append([]string{"install"}, tt.cliArgs...))
			viper.Set("RemoteAPIBaseURL", apiServer.URL)

			var tmpInstallDirSuffix string
			if tt.installDirNotFound {
				tmpInstallDirSuffix = "-nonexistent"
			}
			viper.Set("InstallDir", tempInstallDir+tmpInstallDirSuffix)

			if !tt.installDirNotInPath {
				os.Setenv("PATH", os.ExpandEnv(tempInstallDir+":$PATH"))
			}

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

		err = os.RemoveAll(tempInstallDir)
		if err != nil {
			t.Fatal(err)
		}

		apiServer.Close()
		downloadServer.Close()
	}
}
