package cmd_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
)

func TestAPIDiscoverCmd(t *testing.T) {
	defaultOutRegex := `toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.1.1 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.1.1/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.1.2 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.1.2/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.2.0 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.2.0/toolctl-test-tool.tar.gz
SHA256: .+
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.2.1 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.2.1/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.2.2 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.2.2/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.3.0 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.3.0/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.3.1 ...
URL: .+/(darwin|linux)/(amd|arm)64/0.3.1/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v1.0.0 ...
URL: .+/(darwin|linux)/(amd|arm)64/1.0.0/toolctl-test-tool.tar.gz
HTTP status: 404
toolctl-test-tool (darwin|linux)/(amd|arm)64 v1.0.1 ...
URL: .+/(darwin|linux)/(amd|arm)64/1.0.1/toolctl-test-tool.tar.gz
HTTP status: 404`

	tests := []test{
		{
			name:    "help flag",
			cliArgs: []string{"--help"},
			wantOut: `Discover new versions of supported tools

Usage:
  toolctl api discover [TOOL[@VERSION]...] [flags]

Examples:
  # Discover new versions of all tools
  toolctl discover

  # Discover new versions of a specific tool
  toolctl discover minikube

  # Discover new versions of a specific tool, starting with a specific version
  toolctl discover kubectl@1.20.0

Flags:
      --arch strings   comma-separated list of architectures (default [amd64,arm64])
  -h, --help           help for discover
      --os strings     comma-separated list of operating systems (default [darwin,linux])

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)
`,
		},
		// -------------------------------------------------------------------------
		{
			name: "no cli args",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
				{
					name:                 "toolctl-test-tool",
					version:              "0.2.0",
					onlyOnDownloadServer: true,
					tarGz:                true,
				},
			},
			cliArgs:      []string{},
			wantOutRegex: defaultOutRegex,
			wantFiles: []APIFile{
				{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
				{
					name:                 "toolctl-test-tool",
					version:              "0.2.0",
					onlyOnDownloadServer: true,
					tarGz:                true,
				},
			},
			cliArgs:      []string{"toolctl-test-tool"},
			wantOutRegex: `(?s)`,
			wantFiles: []APIFile{
				{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool with invalid version",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
			},
			cliArgs: []string{"toolctl-test-tool@invalid"},
			wantErr: true,
			wantOut: "Error: Invalid Semantic Version\n",
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool with valid version",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
				{
					name:                 "toolctl-test-tool",
					version:              "0.2.0",
					onlyOnDownloadServer: true,
					tarGz:                true,
				},
			},
			cliArgs: []string{"toolctl-test-tool@0.1.0"},
			wantOutRegex: `toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.1.0 already added
` + defaultOutRegex,
			wantFiles: []APIFile{
				{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		// -------------------------------------------------------------------------
		{
			name: "supported tool with earliest version",
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool",
					version: "0.1.0",
					tarGz:   true,
				},
				{
					name:                 "toolctl-test-tool",
					version:              "0.2.0",
					onlyOnDownloadServer: true,
					tarGz:                true,
				},
			},
			cliArgs: []string{
				"toolctl-test-tool@earliest",
				"--os", runtime.GOOS,
				"--arch", runtime.GOARCH,
			},
			wantOutRegex: `toolctl-test-tool (darwin|linux)/(amd|arm)64 v0.1.0 already added
` + defaultOutRegex,
			wantFiles: []APIFile{
				{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		// -------------------------------------------------------------------------
		{
			name:    "AMD64Default template function",
			cliArgs: []string{"toolctl-test-tool-template-func"},
			supportedTools: []supportedTool{
				{
					name:                    "toolctl-test-tool-template-func",
					version:                 "0.1.0",
					downloadURLTemplatePath: "/v{{.Version}}/{{.Name}}-v{{.Version}}.{{.OS}}{{.Arch | AMD64Default}}",
					tarGz:                   true,
				},
			},
			wantOutRegex: `(?s)URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.darwin.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.darwinarm64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.linux.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.linuxarm64`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "DarwinMacOS and AMD64X64 template functions",
			cliArgs: []string{"toolctl-test-tool-template-func"},
			supportedTools: []supportedTool{
				{
					name:                    "toolctl-test-tool-template-func",
					version:                 "0.1.0",
					downloadURLTemplatePath: "/v{{.Version}}/{{.Name}}-v{{.Version}}.{{.OS | DarwinMacOS}}.{{.Arch | AMD64X64}}",
					tarGz:                   true,
				},
			},
			wantOutRegex: `(?s)URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.macOS.x64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.macOS.arm64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.linux.x64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.linux.arm64`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "AMD64X86_64, ARMUpper and Title template functions",
			cliArgs: []string{"toolctl-test-tool-template-func"},
			supportedTools: []supportedTool{
				{
					name:    "toolctl-test-tool-template-func",
					version: "0.1.0",
					downloadURLTemplatePath: "/v{{.Version}}/" +
						"{{.Name}}-v{{.Version}}.{{.OS | Title}}.{{.Arch | AMD64X86_64 | ARMUpper}}",
					tarGz: true,
				},
			},
			wantOutRegex: `(?s)URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Darwin.x86_64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Darwin.ARM64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Linux.x86_64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Linux.ARM64`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "AMD64Bit and LinuxTitle template functions",
			cliArgs: []string{"toolctl-test-tool-template-func"},
			supportedTools: []supportedTool{
				{
					name:                    "toolctl-test-tool-template-func",
					version:                 "0.1.0",
					downloadURLTemplatePath: "/v{{.Version}}/{{.Name}}-v{{.Version}}.{{.OS | LinuxTitle}}.{{.Arch | AMD64Bit}}",
					tarGz:                   true,
				},
			},
			wantOutRegex: `(?s)URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.darwin.64bit.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.darwin.arm64.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Linux.64bit.+
URL: .+/v0.1.1/toolctl-test-tool-template-func-v0.1.1.Linux.arm64`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "Ignored version",
			cliArgs: []string{"toolctl-test-tool-ignored-versions"},
			supportedTools: []supportedTool{
				{
					name:            "toolctl-test-tool-ignored-versions",
					version:         "0.1.0",
					ignoredVersions: []string{"0.1.1"},
					tarGz:           true,
				},
			},
			wantOutRegex: `v0.1.1 ignored`,
		},
		// -------------------------------------------------------------------------
		{
			name:    "unsupported tool",
			cliArgs: []string{"toolctl-unsupported-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
	}

	for _, tt := range tests {
		localAPIFS, downloadServer, err := setupLocalAPI(tt.supportedTools, true)
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range tt.wantFiles {
			_, err := localAPIFS.Stat(filepath.Join(localAPIBasePath, file.Path))
			if err == nil {
				t.Fatalf("%s: file %s already exists", tt.name, file.Path)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, localAPIFS)
			command.SetArgs(append([]string{"api", "discover"}, tt.cliArgs...))
			viper.Set("LocalAPIBasePath", localAPIBasePath)

			// Redirect Cobra output
			command.SetOut(buf)
			command.SetErr(buf)

			err = command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantOut != "" {
				tt.wantOut = strings.ReplaceAll(tt.wantOut, "{{downloadServerURL}}", downloadServer.URL)
			}
			checkWantOut(t, tt, buf)

			for _, file := range tt.wantFiles {
				_, err := localAPIFS.Stat(filepath.Join(localAPIBasePath, file.Path))
				if err != nil {
					t.Errorf("Error checking file %s: %v", file.Path, err)
				}
			}
		})

		downloadServer.Close()
	}
}
