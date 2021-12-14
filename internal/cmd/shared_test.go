package cmd_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/afero"
	"github.com/toolctl/toolctl/internal/api"
	"github.com/toolctl/toolctl/internal/cmd"
)

const localAPIBasePath = "/toolctl/tools/v0"

type preinstalledTool struct {
	Name         string
	FileContents string
}
type test struct {
	name                        string
	installDirNotFound          bool
	installDirNotInPath         bool
	installDirNotWritable       bool
	preinstalledTools           []preinstalledTool
	preinstalledToolIsSymlinked bool
	cliArgs                     []string
	wantErr                     bool
	wantOut                     string
	wantOutRegex                string
	wantFiles                   APIContents
}

func install(
	t *testing.T, toolctlAPI api.ToolctlAPI, preinstalledTools []preinstalledTool,
	preinstalledToolIsSymlinked bool, originalPathEnv string,
) (tempInstallDir string, err error) {
	tempInstallDir, err = ioutil.TempDir("", "toolctl-test-install-*")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("PATH", os.ExpandEnv(tempInstallDir+":$PATH"))
	if err != nil {
		t.Fatal(err)
	}

	for _, preinstalledTool := range preinstalledTools {
		err = os.WriteFile(
			filepath.Join(tempInstallDir, preinstalledTool.Name),
			[]byte(preinstalledTool.FileContents),
			0755,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	if preinstalledToolIsSymlinked {
		err = os.Rename(
			tempInstallDir+"/toolctl-test-tool",
			tempInstallDir+"/symlinked-toolctl-test-tool",
		)
		if err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(
			tempInstallDir+"/symlinked-toolctl-test-tool",
			tempInstallDir+"/toolctl-test-tool",
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	return
}

func setupRemoteAPI() (
	toolctlAPI api.ToolctlAPI, apiServer *httptest.Server,
	downloadServer *httptest.Server, err error,
) {
	var tarGzSHA256, tarGzInSubdirSHA256 string
	downloadServer, tarGzSHA256, tarGzInSubdirSHA256, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS := afero.NewMemMapFs()

	apiContents := getDefaultAPIContents(
		downloadServer.URL, tarGzSHA256, tarGzInSubdirSHA256,
	)
	for _, f := range apiContents {
		err = afero.WriteFile(localAPIFS, f.Path, []byte(f.Contents), 0644)
		if err != nil {
			return
		}
	}

	apiFileServer := http.FileServer(
		afero.NewHttpFs(localAPIFS).Dir(localAPIBasePath),
	)
	apiServer = httptest.NewServer(apiFileServer)

	toolctlAPI, err = api.NewRemoteAPI(localAPIFS, apiServer.URL)
	if err != nil {
		return
	}

	return
}

func setupLocalAPI() (
	localAPIFS afero.Fs, downloadServer *httptest.Server, err error,
) {
	var tarGzSHA256, tarGzInSubdirSHA256 string
	downloadServer, tarGzSHA256, tarGzInSubdirSHA256, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS = afero.NewMemMapFs()

	apiContents := getDefaultAPIContents(
		downloadServer.URL, tarGzSHA256, tarGzInSubdirSHA256,
	)
	for _, f := range apiContents {
		err = afero.WriteFile(localAPIFS, f.Path, []byte(f.Contents), 0644)
		if err != nil {
			return
		}
	}

	return
}

type APIContents []APIFile

type APIFile struct {
	Path     string
	Contents string
}

func getDefaultAPIContents(
	downloadServerURL string, tarGzFileSHA256 string, tarGzInSubdirSHA256 string,
) APIContents {
	return APIContents{
		// List of supported tools
		APIFile{
			Path: path.Join(localAPIBasePath, "meta.yaml"),
			Contents: `tools:
  - toolctl-test-tool
  - toolctl-another-test-tool
  - toolctl-uninstalled-tool
`,
		},

		// Supported tool
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool/meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool", runtime.GOOS+"-"+runtime.GOARCH, "meta.yaml"),
			Contents: `version:
  earliest: 0.1.0
  latest: 0.1.1
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool", runtime.GOOS+"-"+runtime.GOARCH, "0.1.0.yaml"),
			Contents: `url: ` + downloadServerURL + `/` + runtime.GOOS + `/` + runtime.GOARCH + `/0.1.0/toolctl-test-tool
sha256: 69b2af71462f6deb084b9a38e5ffa2446ab1930232a887c2874d42e81bcc21dd`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool", runtime.GOOS+"-"+runtime.GOARCH, "0.1.1.yaml"),
			Contents: `url: ` + downloadServerURL + `/` + runtime.GOOS + `/` + runtime.GOARCH + `/0.1.1/toolctl-test-tool
sha256: e05dd45f0c922a9ecc659c9f6234159c9820678c0b70d1ca9e48721a379b2143`,
		},

		// Supported tool as .tar.gz
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz/meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}.tar.gz
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz", runtime.GOOS+"-"+runtime.GOARCH, "meta.yaml"),
			Contents: `version:
  earliest: 0.1.0
  latest: 0.1.0
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz", runtime.GOOS+"-"+runtime.GOARCH, "0.1.0.yaml"),
			Contents: fmt.Sprintf(`url: %s/%s/%s/0.1.0/toolctl-test-tool-tar-gz.tar.gz
sha256: %s
`, downloadServerURL, runtime.GOOS, runtime.GOARCH, tarGzFileSHA256),
		},

		// Supported tool as .tar.gz in subdir
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz-subdir/meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}.tar.gz
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz-subdir", runtime.GOOS+"-"+runtime.GOARCH, "meta.yaml"),
			Contents: `version:
  earliest: 0.1.0
  latest: 0.1.0
`,
		},
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-tar-gz-subdir", runtime.GOOS+"-"+runtime.GOARCH, "0.1.0.yaml"),
			Contents: fmt.Sprintf(`url: %s/%s/%s/0.1.0/toolctl-test-tool-tar-gz-subdir.tar.gz
sha256: %s
`, downloadServerURL, runtime.GOOS, runtime.GOARCH, tarGzInSubdirSHA256),
		},

		// Tool that is supported, but not on the current platform
		APIFile{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-unsupported-on-current-platform/meta.yaml"),
			Contents: `description: Test tool unsupported on current OS/Arch
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},

		// Tool with version mismatch
		APIFile{
			Path: path.Join(
				localAPIBasePath, "toolctl-test-tool-version-mismatch/meta.yaml",
			),
			Contents: fmt.Sprintf(`description: toolctl test tool
downloadURLTemplate: %s/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`, downloadServerURL),
		},
		APIFile{
			Path: path.Join(
				localAPIBasePath, "toolctl-test-tool-version-mismatch",
				runtime.GOOS+"-"+runtime.GOARCH, "meta.yaml",
			),
			Contents: `version:
  earliest: 0.1.0
  latest: 0.1.0
`,
		},
		APIFile{
			Path: path.Join(
				localAPIBasePath, "toolctl-test-tool-version-mismatch",
				runtime.GOOS+"-"+runtime.GOARCH, "0.1.0.yaml",
			),
			Contents: fmt.Sprintf(`url: %s/%s/%s/0.1.0/toolctl-test-tool-version-mismatch
sha256: ce04245b5e5ef4aa9b4205b61bedb5e2376a83336b3d72318addbc2c45b553c6
`, downloadServerURL, runtime.GOOS, runtime.GOARCH),
		},
	}
}

func setupDownloadServer() (
	downloadServer *httptest.Server, tarGzSHA256 string,
	tarGzInSubDirSHA256 string, err error,
) {
	downloadFS := afero.NewMemMapFs()

	err = afero.WriteFile(
		downloadFS,
		"/"+runtime.GOOS+"/"+runtime.GOARCH+"/0.1.0/toolctl-test-tool",
		[]byte(`#!/bin/sh
echo "v0.1.0"
`),
		0644,
	)
	if err != nil {
		return
	}

	tarGzSHA256, err = createTarGzTool(downloadFS, false)
	if err != nil {
		return
	}

	tarGzInSubDirSHA256, err = createTarGzTool(downloadFS, true)
	if err != nil {
		return
	}

	err = afero.WriteFile(
		downloadFS,
		"/"+runtime.GOOS+"/"+runtime.GOARCH+"/0.1.1/toolctl-test-tool",
		[]byte(`#!/bin/sh
echo "v0.1.1"
`),
		0644,
	)
	if err != nil {
		return
	}

	err = afero.WriteFile(
		downloadFS,
		"/"+runtime.GOOS+"/"+runtime.GOARCH+"/0.2.0/toolctl-test-tool",
		[]byte(`#!/bin/sh
echo "v0.2.0"
`),
		0644,
	)
	if err != nil {
		return
	}

	err = afero.WriteFile(
		downloadFS,
		"/"+runtime.GOOS+"/"+runtime.GOARCH+"/0.1.0/toolctl-test-tool-version-mismatch",
		[]byte(`#!/bin/sh
echo "v0.2.0"
	`),
		0644,
	)
	if err != nil {
		return
	}

	downloadFileServer := http.FileServer(afero.NewHttpFs(downloadFS).Dir("/"))
	downloadServer = httptest.NewServer(downloadFileServer)

	return
}

func createTarGzTool(downloadFS afero.Fs, inSubdir bool) (sha256 string, err error) {
	filename := "toolctl-test-tool-tar-gz"
	if inSubdir {
		filename += "-subdir"
	}

	tarFilePath := "/" + runtime.GOOS + "/" + runtime.GOARCH + "/0.1.0/" + filename + ".tar"
	tarGzFilePath := tarFilePath + ".gz"
	subdirWithTrailingSlash := ""
	if inSubdir {
		subdirWithTrailingSlash = strings.Replace(path.Base(tarGzFilePath), ".tar.gz", "", 1) + "/"
	}

	tarFile, err := downloadFS.Create(tarFilePath)
	if err != nil {
		return
	}
	tarFileOut, err := downloadFS.OpenFile(tarFile.Name(), os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	tar := archiver.NewTar()
	err = tar.Create(tarFileOut)
	if err != nil {
		return
	}
	inFile, err := downloadFS.Open("/" + runtime.GOOS + "/" + runtime.GOARCH + "/0.1.0/toolctl-test-tool")
	if err != nil {
		return
	}
	inFileStat, err := inFile.Stat()
	if err != nil {
		return
	}
	err = tar.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   inFileStat,
			CustomName: subdirWithTrailingSlash + filename,
		},
		ReadCloser: inFile,
	})
	if err != nil {
		return
	}
	err = tar.Close()
	if err != nil {
		return
	}

	tarFileIn, err := downloadFS.Open(tarFilePath)
	if err != nil {
		return
	}
	tarGzFile, err := downloadFS.Create(tarGzFilePath)
	if err != nil {
		return
	}
	tarGzFileOut, err := downloadFS.OpenFile(tarGzFile.Name(), os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	gz := archiver.NewGz()
	err = gz.Compress(tarFileIn, tarGzFileOut)
	if err != nil {
		return
	}

	tarGzFile, err = downloadFS.Open(tarGzFilePath)
	if err != nil {
		return
	}
	sha256, err = cmd.CalculateSHA256(tarGzFile)
	if err != nil {
		return
	}

	return
}

func TestArgsToTools(t *testing.T) {
	type args struct {
		args           []string
		versionAllowed bool
	}

	tests := []struct {
		name       string
		args       args
		want       []api.Tool
		wantErrStr string
	}{
		{
			name: "should work",
			args: args{
				args:           []string{"test-tool", "other-test-tool@0.1.2"},
				versionAllowed: true,
			},
			want: []api.Tool{
				{
					Name:    "test-tool",
					OS:      runtime.GOOS,
					Arch:    runtime.GOARCH,
					Version: "",
				},
				{
					Name:    "other-test-tool",
					OS:      runtime.GOOS,
					Arch:    runtime.GOARCH,
					Version: "0.1.2",
				},
			},
		},
		{
			name: "version not allowed",
			args: args{
				args:           []string{"test-tool", "test-tool@0.1.2"},
				versionAllowed: false,
			},
			want:       []api.Tool{},
			wantErrStr: "please don't specify a tool version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTools, err := cmd.ArgsToTools(tt.args.args, tt.args.versionAllowed)
			if (err == nil) != (tt.wantErrStr == "") {
				t.Errorf("ArgsToTools() error = %v, wantErr %v", err, tt.wantErrStr)
			}
			if err != nil && err.Error() != tt.wantErrStr {
				t.Errorf("ArgsToTools() error = %v, wantErr %v", err, tt.wantErrStr)
			}
			if !cmp.Equal(gotTools, tt.want) {
				t.Errorf("ArgsToTools() = %v, want %v", gotTools, tt.want)
			}
		})
	}
}
