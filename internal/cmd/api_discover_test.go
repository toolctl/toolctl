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
	tests := []test{
		{
			name:    "no cli args",
			cliArgs: []string{},
			wantErr: true,
			wantOut: `Error: no tool specified
Usage:
  toolctl api discover TOOL[@VERSION]... [flags]

Examples:
  # Discover new versions of kubectl
  toolctl discover kubectl

  # Discover new versions of kubectl, starting with v1.20.0
  toolctl discover kubectl@1.20.0

Flags:
      --arch strings   comma-separated list of architectures (default [amd64,arm64])
  -h, --help           help for discover
      --os strings     comma-separated list of operating systems (default [darwin,linux])

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)

`,
		},
		{
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			wantOut: `toolctl-test-tool darwin/amd64 v0.1.2 ...
URL: {{downloadServerURL}}/darwin/amd64/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.1.3 ...
URL: {{downloadServerURL}}/darwin/amd64/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.2.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool darwin/amd64 v0.2.1 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.2.2 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.3.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.4.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v1.0.0 ...
URL: {{downloadServerURL}}/darwin/amd64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v2.0.0 ...
URL: {{downloadServerURL}}/darwin/amd64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.0.1 ...
URL: {{downloadServerURL}}/darwin/arm64/0.0.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.0.2 ...
URL: {{downloadServerURL}}/darwin/arm64/0.0.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.1.0 ...
URL: {{downloadServerURL}}/darwin/arm64/0.1.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.2.0 ...
URL: {{downloadServerURL}}/darwin/arm64/0.2.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v1.0.0 ...
URL: {{downloadServerURL}}/darwin/arm64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v2.0.0 ...
URL: {{downloadServerURL}}/darwin/arm64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.1.2 ...
URL: {{downloadServerURL}}/linux/amd64/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.1.3 ...
URL: {{downloadServerURL}}/linux/amd64/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.2.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool linux/amd64 v0.2.1 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.2.2 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.3.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.4.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v1.0.0 ...
URL: {{downloadServerURL}}/linux/amd64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v2.0.0 ...
URL: {{downloadServerURL}}/linux/amd64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.0.1 ...
URL: {{downloadServerURL}}/linux/arm64/0.0.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.0.2 ...
URL: {{downloadServerURL}}/linux/arm64/0.0.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.1.0 ...
URL: {{downloadServerURL}}/linux/arm64/0.1.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.2.0 ...
URL: {{downloadServerURL}}/linux/arm64/0.2.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v1.0.0 ...
URL: {{downloadServerURL}}/linux/arm64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v2.0.0 ...
URL: {{downloadServerURL}}/linux/arm64/2.0.0/toolctl-test-tool
HTTP status: 404
`,
			wantFiles: APIContents{
				APIFile{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		{
			name:    "supported tool with version",
			cliArgs: []string{"toolctl-test-tool@0.1.0"},
			wantOut: `toolctl-test-tool darwin/amd64 v0.1.0 already added
toolctl-test-tool darwin/amd64 v0.1.1 already added
toolctl-test-tool darwin/amd64 v0.1.2 ...
URL: {{downloadServerURL}}/darwin/amd64/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.1.3 ...
URL: {{downloadServerURL}}/darwin/amd64/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.2.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool darwin/amd64 v0.2.1 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.2.2 ...
URL: {{downloadServerURL}}/darwin/amd64/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.3.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v0.4.0 ...
URL: {{downloadServerURL}}/darwin/amd64/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v1.0.0 ...
URL: {{downloadServerURL}}/darwin/amd64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/amd64 v2.0.0 ...
URL: {{downloadServerURL}}/darwin/amd64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.1.0 ...
URL: {{downloadServerURL}}/darwin/arm64/0.1.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.1.1 ...
URL: {{downloadServerURL}}/darwin/arm64/0.1.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.2.0 ...
URL: {{downloadServerURL}}/darwin/arm64/0.2.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v0.3.0 ...
URL: {{downloadServerURL}}/darwin/arm64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v1.0.0 ...
URL: {{downloadServerURL}}/darwin/arm64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool darwin/arm64 v2.0.0 ...
URL: {{downloadServerURL}}/darwin/arm64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.1.0 already added
toolctl-test-tool linux/amd64 v0.1.1 already added
toolctl-test-tool linux/amd64 v0.1.2 ...
URL: {{downloadServerURL}}/linux/amd64/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.1.3 ...
URL: {{downloadServerURL}}/linux/amd64/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.2.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool linux/amd64 v0.2.1 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.2.2 ...
URL: {{downloadServerURL}}/linux/amd64/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.3.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v0.4.0 ...
URL: {{downloadServerURL}}/linux/amd64/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v1.0.0 ...
URL: {{downloadServerURL}}/linux/amd64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/amd64 v2.0.0 ...
URL: {{downloadServerURL}}/linux/amd64/2.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.1.0 ...
URL: {{downloadServerURL}}/linux/arm64/0.1.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.1.1 ...
URL: {{downloadServerURL}}/linux/arm64/0.1.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.2.0 ...
URL: {{downloadServerURL}}/linux/arm64/0.2.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v0.3.0 ...
URL: {{downloadServerURL}}/linux/arm64/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v1.0.0 ...
URL: {{downloadServerURL}}/linux/arm64/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool linux/arm64 v2.0.0 ...
URL: {{downloadServerURL}}/linux/arm64/2.0.0/toolctl-test-tool
HTTP status: 404
`,
			wantFiles: APIContents{
				APIFile{
					Path: fmt.Sprintf(
						"toolctl-test-tool/%s-%s/0.2.0.yaml", runtime.GOOS, runtime.GOARCH,
					),
				},
			},
		},
		{
			name:    "unsupported tool",
			cliArgs: []string{"toolctl-unsupported-test-tool"},
			wantErr: true,
			wantOut: `Error: toolctl-unsupported-test-tool could not be found
`,
		},
	}

	for _, tt := range tests {
		localAPIFS, downloadServer, err := setupLocalAPI()
		if err != nil {
			t.Fatal(err)
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
