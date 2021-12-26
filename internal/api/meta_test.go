package api_test

import (
	"errors"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/toolctl/toolctl/internal/api"
)

func TestGetMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
	}

	tests := []struct {
		name        string
		apiContents apiContents
		args        args
		want        api.Meta
		wantErr     error
	}{
		{
			name: "should work",
			apiContents: []apiFile{
				{
					Path: filepath.Join(localAPIBasePath, "meta.yaml"),
					Contents: `tools:
- toolctl-test-tool
- toolctl-other-test-tool
`,
				},
			},
			want: api.Meta{
				Tools: []string{
					"toolctl-test-tool",
					"toolctl-other-test-tool",
				},
			},
		},
		{
			name:    "should fail if meta.yaml is missing",
			wantErr: api.NotFoundError{},
		},
	}

	for _, tt := range tests {
		for _, apiLocation := range []api.Location{api.Remote, api.Local} {
			toolctlAPI, apiServer, err := setupTest(apiLocation, tt.apiContents)
			if err != nil {
				t.Fatal(err)
			}

			t.Run(tt.name, func(t *testing.T) {
				got, err := api.GetMeta(toolctlAPI)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetMeta() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetMeta() = %v, want %v", got, tt.want)
				}
			})

			if apiLocation == api.Remote {
				apiServer.Close()
			}
		}
	}
}

func TestSaveMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		meta api.Meta
	}
	tests := []struct {
		name            string
		apiLocation     api.Location
		args            args
		wantAPIContents apiContents
		wantErrStr      string
	}{
		{
			name:        "save with local API",
			apiLocation: api.Local,
			args: args{
				meta: api.Meta{
					Tools: []string{
						"toolctl-test-tool",
						"toolctl-other-test-tool",
					},
				},
			},
			wantAPIContents: apiContents{
				{
					Path: path.Join(localAPIBasePath, "meta.yaml"),
					Contents: `tools:
  - toolctl-test-tool
  - toolctl-other-test-tool
`,
				},
			},
		},
		{
			name:        "save with remote API",
			apiLocation: api.Remote,
			args: args{
				meta: api.Meta{
					Tools: []string{
						"toolctl-test-tool",
					},
				},
			},
			wantErrStr: "not implemented",
		},
	}
	for _, tt := range tests {
		toolctlAPI, apiServer, err := setupTest(tt.apiLocation, apiContents{})
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			err := api.SaveMeta(toolctlAPI, tt.args.meta)
			if (err == nil) != (tt.wantErrStr == "") {
				t.Errorf("SavePlatformMeta() error = %v, wantErr = %v", err, tt.wantErrStr)
			}
			if err != nil && tt.wantErrStr != "" {
				if err.Error() != tt.wantErrStr {
					t.Errorf("SavePlatformMeta() error = %v, wantErr = %v", err, tt.wantErrStr)
				}
			}
		})

		if tt.apiLocation == api.Remote {
			apiServer.Close()
		}
	}
}

func TestGetToolMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		tool api.Tool
	}

	tests := []struct {
		name        string
		apiContents apiContents
		args        args
		want        api.ToolMeta
		wantErr     error
	}{
		{
			name: "supported tool",
			apiContents: apiContents{
				apiFile{
					Path:     filepath.Join(localAPIBasePath, "toolctl-test-tool/meta.yaml"),
					Contents: "downloadURLTemplate: https://localhost/{{.OS}}/{{.Arch}}/{{.Version}}/{{.ToolName}}",
				},
			},
			args: args{
				tool: api.Tool{
					Name: "toolctl-test-tool",
				},
			},
			want: api.ToolMeta{
				DownloadURLTemplate: "https://localhost/{{.OS}}/{{.Arch}}/{{.Version}}/{{.ToolName}}",
			},
			wantErr: nil,
		},
		{
			name: "unsupported tool",
			apiContents: apiContents{
				apiFile{
					Path:     localAPIBasePath + "/toolctl-test-tool/meta.yaml",
					Contents: "downloadURLTemplate: https://localhost/{{.OS}}/{{.Arch}}/{{.Version}}/{{.ToolName}}",
				},
			},
			args: args{
				tool: api.Tool{
					Name: "toolctl-unsupported-test-tool",
				},
			},
			wantErr: api.NotFoundError{},
		},
	}

	for _, tt := range tests {
		for _, apiLocation := range []api.Location{api.Remote, api.Local} {
			toolctlAPI, apiServer, err := setupTest(apiLocation, tt.apiContents)
			if err != nil {
				t.Fatal(err)
			}

			t.Run(tt.name, func(t *testing.T) {
				got, err := api.GetToolMeta(toolctlAPI, tt.args.tool)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetMeta() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetMeta() = %v, want %v", got, tt.want)
				}
			})

			if apiLocation == api.Remote {
				apiServer.Close()
			}
		}
	}
}

func TestGetToolPlatformMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		tool api.Tool
	}
	tests := []struct {
		name        string
		apiContents apiContents
		args        args
		want        api.ToolPlatformMeta
		wantErr     bool
	}{
		{
			name: "found",
			apiContents: apiContents{
				{
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
			want: api.ToolPlatformMeta{
				Version: api.ToolPlatformMetaVersion{
					Earliest: "1.0.0",
					Latest:   "1.3.2",
				},
			},
		},
		{
			name:        "could not be found",
			apiContents: apiContents{},
			args: args{
				tool: api.Tool{
					Name: "toolctl-unsupported-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
			},
			want:    api.ToolPlatformMeta{},
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
				got, err := api.GetToolPlatformMeta(toolctlAPI, tt.args.tool)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetPlatformMeta() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetPlatformMeta() = %v, want %v", got, tt.want)
				}
			})

			if apiLocation == api.Remote {
				apiServer.Close()
			}
		}
	}
}

func TestSaveToolPlatformMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		tool api.Tool
		meta api.ToolPlatformMeta
	}
	tests := []struct {
		name            string
		apiLocation     api.Location
		args            args
		wantAPIContents apiContents
		wantErrStr      string
	}{
		{
			name:        "save with local API",
			apiLocation: api.Local,
			args: args{
				tool: api.Tool{
					Name: "toolctl-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
				meta: api.ToolPlatformMeta{
					Version: api.ToolPlatformMetaVersion{
						Earliest: "0.1.0",
						Latest:   "1.3.2",
					},
				},
			},
			wantAPIContents: apiContents{
				{
					Path: path.Join(localAPIBasePath, "toolctl-test-tool/darwin-amd64/meta.yaml"),
					Contents: `earliest: 0.1.0
latest: 1.3.2
`,
				},
			},
		},
		{
			name:        "save with remote API",
			apiLocation: api.Remote,
			args: args{
				tool: api.Tool{
					Name: "toolctl-test-tool",
					OS:   "darwin",
					Arch: "amd64",
				},
				meta: api.ToolPlatformMeta{
					Version: api.ToolPlatformMetaVersion{
						Earliest: "0.1.0",
						Latest:   "1.3.2",
					},
				},
			},
			wantErrStr: "not implemented",
		},
	}
	for _, tt := range tests {
		toolctlAPI, apiServer, err := setupTest(tt.apiLocation, apiContents{})
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			err := api.SaveToolPlatformMeta(toolctlAPI, tt.args.tool, tt.args.meta)
			if (err == nil) != (tt.wantErrStr == "") {
				t.Errorf("SavePlatformMeta() error = %v, wantErr = %v", err, tt.wantErrStr)
			}
			if err != nil && tt.wantErrStr != "" {
				if err.Error() != tt.wantErrStr {
					t.Errorf("SavePlatformMeta() error = %v, wantErr = %v", err, tt.wantErrStr)
				}
			}
		})

		if tt.apiLocation == api.Remote {
			apiServer.Close()
		}
	}
}

func TestGetToolPlatformVersionMeta(t *testing.T) {
	type args struct {
		// toolctlAPI is created for each test case
		tool api.Tool
	}
	tests := []struct {
		name                string
		apiContents         apiContents
		args                args
		wantPlatformVersion api.ToolPlatformVersionMeta
		wantErr             error
	}{
		{
			name: "supported tool",
			apiContents: apiContents{
				{
					Path: path.Join(localAPIBasePath, "toolctl-test-tool/darwin-amd64/1.0.0.yaml"),
					Contents: `
url: https://localhost/release/v1.0.0/bin/darwin/amd64/toolctl-test-tool
sha256: cb3174cf3910a0d711a61059363aad6a30b7dcc1125be8027f20907a6612bf24
`,
				},
			},
			args: args{
				tool: api.Tool{
					Name:    "toolctl-test-tool",
					OS:      "darwin",
					Arch:    "amd64",
					Version: "1.0.0",
				},
			},
			wantPlatformVersion: api.ToolPlatformVersionMeta{
				URL:    "https://localhost/release/v1.0.0/bin/darwin/amd64/toolctl-test-tool",
				SHA256: "cb3174cf3910a0d711a61059363aad6a30b7dcc1125be8027f20907a6612bf24",
			},
		},
		{
			name:        "unsupported version",
			apiContents: apiContents{},
			args: args{
				tool: api.Tool{
					Name:    "toolctl-test-tool",
					OS:      "darwin",
					Arch:    "amd64",
					Version: "2.0.0",
				},
			},
			wantErr: api.NotFoundError{},
		},
		{
			name:        "unsupported tool",
			apiContents: apiContents{},
			args: args{
				tool: api.Tool{
					Name:    "toolctl-unsupported-test-tool",
					OS:      "darwin",
					Arch:    "amd64",
					Version: "1.0.0",
				},
			},
			wantErr: api.NotFoundError{},
		},
	}
	for _, tt := range tests {
		for _, apiLocation := range []api.Location{api.Remote, api.Local} {
			toolctlAPI, apiServer, err := setupTest(apiLocation, tt.apiContents)
			if err != nil {
				t.Fatal(err)
			}

			t.Run(tt.name, func(t *testing.T) {
				gotPlatformVersion, err := api.GetToolPlatformVersionMeta(toolctlAPI, tt.args.tool)

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetPlatformVersion() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !cmp.Equal(gotPlatformVersion, tt.wantPlatformVersion) {
					t.Errorf("GetPlatformVersion() = %v, want %v", gotPlatformVersion, tt.wantPlatformVersion)
				}
			})

			if apiLocation == api.Remote {
				apiServer.Close()
			}
		}
	}
}

func TestSaveToolPlatformVersionMeta(t *testing.T) {
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
			if err := api.SaveToolPlatformVersionMeta(toolctlAPI, tt.args.tool, tt.args.meta); (err != nil) != tt.wantErr {
				t.Errorf("SaveVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		if tt.apiLocation == api.Remote {
			apiServer.Close()
		}
	}
}
