package api_test

import (
	"path"
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/toolctl/toolctl/internal/api"
)

func TestGetLatestVersion(t *testing.T) {
	type args struct {
		tool api.Tool
		// toolctlAPI is created for each test case
	}
	tests := []struct {
		name        string
		apiContents apiContents
		args        args
		want        *semver.Version
		wantErr     bool
	}{
		{
			name: "supported tool",
			apiContents: apiContents{
				apiFile{
					Path: path.Join(localAPIBasePath, "meta.yaml"),
				},
				apiFile{
					Path: path.Join(localAPIBasePath, "toolctl-test-tool/darwin-amd64/meta.yaml"),
					Contents: `version:
  earliest: 1.0.0
  latest: 1.3.2
`,
				},
			},
			args: args{
				tool: api.Tool{
					Name: "toolctl-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
			},
			want: semver.MustParse("1.3.2"),
		},
		{
			name: "unsupported tool",
			args: args{
				tool: api.Tool{
					Name: "toolctl-unsupported-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		for _, apiLocation := range []api.Location{api.Remote, api.Local} {
			toolctlAPI, apiServer, err := setupTest(apiLocation, tt.apiContents)
			if err != nil {
				t.Fatal(err)
			}

			t.Run(tt.name, func(t *testing.T) {
				got, err := api.GetLatestVersion(toolctlAPI, tt.args.tool)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetLatestVersion() = %v, want %v", got, tt.want)
				}
			})

			if apiLocation == api.Remote {
				apiServer.Close()
			}
		}
	}
}
