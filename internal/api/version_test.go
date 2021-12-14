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
		// api is created for each test case
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

func TestSaveVersion(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		tool api.Tool
		meta api.ToolPlatformVersionMeta
	}

	tests := []struct {
		name        string
		apiLocation api.Location
		args        args
		wantErr     bool
	}{
		{
			name:        "local",
			apiLocation: api.Local,
			args: args{
				tool: api.Tool{
					Name: "toolctl-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
				meta: api.ToolPlatformVersionMeta{
					URL:    "http://localhost/toolctl-test-tool/darwin-amd64/toolctl-test-tool-1.0.0.tar.gz",
					SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
		},
		{
			name:        "remote",
			apiLocation: api.Remote,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		toolctlAPI, apiServer, err := setupTest(tt.apiLocation, apiContents{})
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := api.SaveVersion(toolctlAPI, tt.args.tool, tt.args.meta); (err != nil) != tt.wantErr {
				t.Errorf("SaveVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		if tt.apiLocation == api.Remote {
			apiServer.Close()
		}
	}
}
