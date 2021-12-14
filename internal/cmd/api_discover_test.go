package cmd_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
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
  -h, --help   help for discover

Global Flags:
      --config string   path of the config file (default is $HOME/.config/toolctl/config.yaml)

`,
		},
		{
			name:    "supported tool",
			cliArgs: []string{"toolctl-test-tool"},
			wantOutRegex: `^toolctl-test-tool v0.1.2 ...
URL: http://127.0.0.1:.+/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.1.3 ...
URL: http://127.0.0.1:.+/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.2.0 ...
URL: http://127.0.0.1:.+/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool v0.2.1 ...
URL: http://127.0.0.1:.+/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.2.2 ...
URL: http://127.0.0.1:.+/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.3.0 ...
URL: http://127.0.0.1:.+/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.4.0 ...
URL: http://127.0.0.1:.+/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v1.0.0 ...
URL: http://127.0.0.1:.+/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v2.0.0 ...
URL: http://127.0.0.1:.+/2.0.0/toolctl-test-tool
HTTP status: 404
$`,
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
			wantOutRegex: `^toolctl-test-tool v0.1.0 already added
toolctl-test-tool v0.1.1 already added
toolctl-test-tool v0.1.2 ...
URL: http://127.0.0.1:.+/0.1.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.1.3 ...
URL: http://127.0.0.1:.+/0.1.3/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.2.0 ...
URL: http://127.0.0.1:.+/0.2.0/toolctl-test-tool
SHA256: 92100d959b50115ff3760255480a1ecb6f8558d26f7757cd46b45223f13ac6f1
toolctl-test-tool v0.2.1 ...
URL: http://127.0.0.1:.+/0.2.1/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.2.2 ...
URL: http://127.0.0.1:.+/0.2.2/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.3.0 ...
URL: http://127.0.0.1:.+/0.3.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v0.4.0 ...
URL: http://127.0.0.1:.+/0.4.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v1.0.0 ...
URL: http://127.0.0.1:.+/1.0.0/toolctl-test-tool
HTTP status: 404
toolctl-test-tool v2.0.0 ...
URL: http://127.0.0.1:.+/2.0.0/toolctl-test-tool
HTTP status: 404
$`,
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
