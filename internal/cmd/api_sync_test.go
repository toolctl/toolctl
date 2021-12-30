package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/cmd"
)

func TestAPISyncCmd(t *testing.T) {
	tests := []test{
		{
			name:         "should work",
			cliArgs:      []string{},
			wantOutRegex: "^$",
			wantFiles: []APIFile{
				{
					Path: "meta.yaml",
					Contents: `tools:
  - toolctl-test-tool
  - toolctl-test-tool-unsupported-on-current-platform
`,
				},
			},
		},
	}

	for _, tt := range tests {
		localAPIFS, downloadServer, err := setupLocalAPI(tt.supportedTools)
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, localAPIFS)
			command.SetArgs(append([]string{"api", "sync"}, tt.cliArgs...))
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
				if file.Contents != "" {
					fileContents, err := afero.ReadFile(localAPIFS, filepath.Join(localAPIBasePath, file.Path))
					if err != nil {
						t.Errorf("Error reading file %s: %v", file.Path, err)
					}
					if string(fileContents) != file.Contents {
						t.Errorf("File %s contents = %s, want %s", file.Path, string(fileContents), file.Contents)
					}
				}
			}
		})

		downloadServer.Close()
	}
}
