package cmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/api"
	"github.com/toolctl/toolctl/internal/cmd"
)

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
			gotTools, err := cmd.ArgsToTools(tt.args.args, runtime.GOOS, runtime.GOARCH, tt.args.versionAllowed)
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

// -----------------------------------------------------------------------------
// Shared test helpers
// -----------------------------------------------------------------------------

const localAPIBasePath = "/toolctl/tools/v0"

type preinstalledTool struct {
	name         string
	fileContents string
}

type supportedTool struct {
	name            string
	version         string
	tarGz           bool
	tarGzSubdir     string
	tarGzBinaryName string
}

type test struct {
	name                        string
	installDirNotFound          bool
	installDirNotInPath         bool
	installDirNotWritable       bool
	installDirNotPreinstallDir  bool
	supportedTools              []supportedTool
	preinstalledTools           []preinstalledTool
	preinstalledToolIsSymlinked bool
	cliArgs                     []string
	wantErr                     bool
	wantOut                     string
	wantOutRegex                string
	wantFiles                   []APIFile
}

func setupPreinstallTempDir(
	t *testing.T, tt test, toolctlAPI api.ToolctlAPI, originalPathEnv string,
) (preinstallTempDir string) {
	preinstallTempDir, err := ioutil.TempDir("", "toolctl-test-install-*")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("PATH", os.ExpandEnv(preinstallTempDir+":$PATH"))
	if err != nil {
		t.Fatal(err)
	}

	for _, preinstalledTool := range tt.preinstalledTools {
		err = os.WriteFile(
			filepath.Join(preinstallTempDir, preinstalledTool.name),
			[]byte(preinstalledTool.fileContents),
			0755,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	if tt.preinstalledToolIsSymlinked {
		err = os.Rename(
			preinstallTempDir+"/toolctl-test-tool",
			preinstallTempDir+"/symlinked-toolctl-test-tool",
		)
		if err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(
			preinstallTempDir+"/symlinked-toolctl-test-tool",
			preinstallTempDir+"/toolctl-test-tool",
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	return
}

func setupRemoteAPI(supportedTools []supportedTool) (
	toolctlAPI api.ToolctlAPI, apiServer *httptest.Server,
	downloadServer *httptest.Server, err error,
) {
	var downloadServerFS afero.Fs
	downloadServer, downloadServerFS, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS := afero.NewMemMapFs()

	apiFiles := getDefaultAPIFiles(downloadServer.URL)

	for _, supportedTool := range supportedTools {
		var sha256 string
		sha256, err = supportedToolToDownloadFile(downloadServerFS, supportedTool)
		if err != nil {
			return
		}
		apiFiles = append(
			apiFiles,
			supportedToolToAPIContents(supportedTool, downloadServer.URL, sha256)...,
		)
	}

	for _, f := range apiFiles {
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
	downloadServer, _, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS = afero.NewMemMapFs()

	apiFiles := getDefaultAPIFiles(downloadServer.URL)
	for _, f := range apiFiles {
		err = afero.WriteFile(localAPIFS, f.Path, []byte(f.Contents), 0644)
		if err != nil {
			return
		}
	}

	return
}

type APIFile struct {
	Path     string
	Contents string
}

func getDefaultAPIFiles(downloadServerURL string) []APIFile {
	apiFiles := []APIFile{
		// List of supported tools
		{
			Path: path.Join(localAPIBasePath, "meta.yaml"),
			Contents: `tools:
  - toolctl-test-tool
  - toolctl-another-test-tool
  - toolctl-uninstalled-tool
`,
		},

		// Supported tool
		{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool/meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},

		// Tool that is supported, but not on the current platform
		{
			Path: path.Join(localAPIBasePath, "toolctl-test-tool-unsupported-on-current-platform/meta.yaml"),
			Contents: `description: Test tool unsupported on current OS/Arch
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},

		// Tool with version mismatch
		{
			Path: path.Join(
				localAPIBasePath, "toolctl-test-tool-version-mismatch/meta.yaml",
			),
			Contents: fmt.Sprintf(`description: toolctl test tool
downloadURLTemplate: %s/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}
homepage: https://toolctl.io/
versionArgs: [version, --short]
`, downloadServerURL),
		},
	}

	for _, os := range []string{"darwin", "linux"} {
		apiFiles = append(apiFiles,
			// Supported tool
			APIFile{
				Path: path.Join(localAPIBasePath, "toolctl-test-tool", os+"-amd64", "meta.yaml"),
				Contents: `version:
  earliest: 0.1.0
  latest: 0.1.1
`,
			},
			APIFile{
				Path: path.Join(localAPIBasePath, "toolctl-test-tool", os+"-amd64", "0.1.0.yaml"),
				Contents: `url: ` + downloadServerURL + `/` + os + `/amd64/0.1.0/toolctl-test-tool
sha256: 69b2af71462f6deb084b9a38e5ffa2446ab1930232a887c2874d42e81bcc21dd`,
			},
			APIFile{
				Path: path.Join(localAPIBasePath, "toolctl-test-tool", os+"-amd64", "0.1.1.yaml"),
				Contents: `url: ` + downloadServerURL + `/` + os + `/amd64/0.1.1/toolctl-test-tool
sha256: e05dd45f0c922a9ecc659c9f6234159c9820678c0b70d1ca9e48721a379b2143`,
			},

			// Tool with version mismatch
			APIFile{
				Path: path.Join(
					localAPIBasePath, "toolctl-test-tool-version-mismatch",
					os+"-amd64", "meta.yaml",
				),
				Contents: `version:
  earliest: 0.1.0
  latest: 0.1.0
`,
			},
			APIFile{
				Path: path.Join(
					localAPIBasePath, "toolctl-test-tool-version-mismatch",
					os+"-amd64", "0.1.0.yaml",
				),
				Contents: fmt.Sprintf(`url: %s/%s/%s/0.1.0/toolctl-test-tool-version-mismatch
sha256: ce04245b5e5ef4aa9b4205b61bedb5e2376a83336b3d72318addbc2c45b553c6
`, downloadServerURL, os, "amd64"),
			},
		)
	}

	return apiFiles
}

func setupDownloadServer() (
	downloadServer *httptest.Server, downloadServerFS afero.Fs, err error,
) {
	downloadServerFS = afero.NewMemMapFs()

	for _, os := range []string{"darwin", "linux"} {
		err = afero.WriteFile(
			downloadServerFS,
			"/"+os+"/amd64/0.1.0/toolctl-test-tool",
			[]byte(`#!/bin/sh
echo "v0.1.0"
`),
			0644,
		)
		if err != nil {
			return
		}

		err = afero.WriteFile(
			downloadServerFS,
			"/"+os+"/amd64/0.1.1/toolctl-test-tool",
			[]byte(`#!/bin/sh
echo "v0.1.1"
`),
			0644,
		)
		if err != nil {
			return
		}

		err = afero.WriteFile(
			downloadServerFS,
			"/"+os+"/amd64/0.2.0/toolctl-test-tool",
			[]byte(`#!/bin/sh
echo "v0.2.0"
`),
			0644,
		)
		if err != nil {
			return
		}

		err = afero.WriteFile(
			downloadServerFS,
			"/"+os+"/amd64/0.1.0/toolctl-test-tool-version-mismatch",
			[]byte(`#!/bin/sh
echo "v0.2.0"
	`),
			0644,
		)
		if err != nil {
			return
		}
	}

	downloadFileServer := http.FileServer(
		afero.NewHttpFs(downloadServerFS).Dir("/"),
	)
	downloadServer = httptest.NewServer(downloadFileServer)

	return
}

func calculateSHA256(
	downloadFS afero.Fs, tarGzFilePath string,
) (sha256 string, err error) {
	tarGzFile, err := downloadFS.Open(tarGzFilePath)
	if err != nil {
		return
	}
	defer tarGzFile.Close()

	sha256, err = cmd.CalculateSHA256(tarGzFile)
	if err != nil {
		return
	}

	return
}

func checkWantOut(t *testing.T, tt test, buf *bytes.Buffer) {
	if tt.wantOut == "" && tt.wantOutRegex == "" {
		t.Fatalf("Either wantOut or wantOutRegex must be set")
	}
	if tt.wantOut != "" && tt.wantOutRegex != "" {
		t.Fatalf("wantOut and wantOutRegex cannot be set at the same time")
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
			t.Errorf(
				"Error matching regex: %v, output: %s",
				tt.wantOutRegex, buf.String(),
			)
		}
	}
}

func supportedToolToDownloadFile(
	downloadServerFS afero.Fs, supportedTool supportedTool,
) (sha256 string, err error) {
	if !supportedTool.tarGz {
		err = fmt.Errorf("Only tar.gz supported for now")
		return
	}

	if supportedTool.tarGzBinaryName == "" {
		supportedTool.tarGzBinaryName = supportedTool.name
	}

	filePath, err := createBinaryFile(downloadServerFS, supportedTool)
	if err != nil {
		return
	}

	tarFilePath, err := createTarFile(downloadServerFS, filePath, supportedTool)
	if err != nil {
		return
	}

	tarGzFilePath, err := createTarGzFile(tarFilePath, downloadServerFS)
	if err != nil {
		return
	}

	sha256, err = calculateSHA256(downloadServerFS, tarGzFilePath)
	if err != nil {
		return
	}

	return
}

func createTarGzFile(
	tarFilePath string, downloadServerFS afero.Fs,
) (tarGzFilePath string, err error) {
	tarGzFilePath = tarFilePath + ".gz"

	tarFileIn, err := downloadServerFS.Open(tarFilePath)
	if err != nil {
		return
	}

	tarGzFile, err := downloadServerFS.Create(tarGzFilePath)
	if err != nil {
		return
	}

	tarGzFileOut, err := downloadServerFS.OpenFile(
		tarGzFile.Name(), os.O_WRONLY, 0644,
	)
	if err != nil {
		return
	}

	err = archiver.NewGz().Compress(tarFileIn, tarGzFileOut)

	return
}

func createTarFile(
	downloadServerFS afero.Fs, filePath string, supportedTool supportedTool,
) (tarFilePath string, err error) {
	tarFilePath = filePath + ".tar"
	tarFile, err := downloadServerFS.Create(tarFilePath)
	if err != nil {
		return
	}
	defer tarFile.Close()

	tarFileOut, err := downloadServerFS.OpenFile(
		tarFile.Name(), os.O_WRONLY, 0644,
	)
	if err != nil {
		return
	}

	tar := archiver.NewTar()
	err = tar.Create(tarFileOut)
	if err != nil {
		return
	}

	inFile, err := downloadServerFS.Open(filePath)
	if err != nil {
		return
	}

	inFileStat, err := inFile.Stat()
	if err != nil {
		return
	}

	err = tar.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo: inFileStat,
			CustomName: filepath.Join(
				supportedTool.tarGzSubdir, supportedTool.tarGzBinaryName),
		},
		ReadCloser: inFile,
	})
	if err != nil {
		return
	}

	return
}

func createBinaryFile(
	downloadServerFS afero.Fs, supportedTool supportedTool,
) (filePath string, err error) {
	filePath = "/" + filepath.Join(
		runtime.GOOS, runtime.GOARCH, supportedTool.version, supportedTool.name,
	)
	err = afero.WriteFile(
		downloadServerFS,
		filePath,
		[]byte(`#!/bin/sh
echo v`+supportedTool.version+`
`),
		0644,
	)
	return
}

func supportedToolToAPIContents(
	supportedTool supportedTool, downloadServerURL string, sha256 string,
) (apiFiles []APIFile) {
	return []APIFile{
		{
			Path: path.Join(localAPIBasePath, supportedTool.name, "meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + `/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}.tar.gz
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},
		{
			Path: path.Join(
				localAPIBasePath, supportedTool.name, runtime.GOOS+"-"+runtime.GOARCH,
				"meta.yaml",
			),
			Contents: fmt.Sprintf(`version:
  earliest: %s
  latest: %s
`, supportedTool.version, supportedTool.version),
		},
		{
			Path: path.Join(
				localAPIBasePath, supportedTool.name, runtime.GOOS+"-"+runtime.GOARCH,
				supportedTool.version+".yaml",
			),
			Contents: fmt.Sprintf(
				`url: %s/%s/%s/%s/%s.tar.gz
sha256: %s
`,
				downloadServerURL, runtime.GOOS, runtime.GOARCH, supportedTool.version,
				supportedTool.name, sha256,
			),
		},
	}
}

func runInstallUpgradeTests(
	t *testing.T, tests []test, installOrUpgrade string,
) {
	originalPathEnv := os.Getenv("PATH")

	for _, tt := range tests {
		toolctlAPI, apiServer, downloadServer, err := setupRemoteAPI(
			tt.supportedTools,
		)
		if err != nil {
			t.Fatal(err)
		}

		installTempDir, err := ioutil.TempDir("", "toolctl-test-install-*")
		if err != nil {
			t.Fatal(err)
		}

		var preinstallTempDir string
		if !cmp.Equal(tt.preinstalledTools, []preinstalledTool{}) {
			preinstallTempDir = setupPreinstallTempDir(
				t, tt, toolctlAPI, originalPathEnv,
			)
		}

		if !tt.installDirNotPreinstallDir && !tt.installDirNotInPath {
			installTempDir = preinstallTempDir
		} else {
			installTempDir = setupInstallTempDir(t, tt)
		}

		if tt.installDirNotWritable {
			err = os.Chmod(installTempDir, 0500)
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			command := cmd.NewRootCmd(buf, toolctlAPI.GetLocalAPIFS())
			command.SetArgs(append([]string{installOrUpgrade}, tt.cliArgs...))
			viper.Set("RemoteAPIBaseURL", apiServer.URL)

			var tmpInstallDirSuffix string
			if tt.installDirNotFound {
				tmpInstallDirSuffix = "-nonexistent"
			}
			viper.Set("InstallDir", installTempDir+tmpInstallDirSuffix)

			// Redirect Cobra output to a buffer
			command.SetOut(buf)
			command.SetErr(buf)

			err := command.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}

			checkWantOut(t, tt, buf)
		})

		os.Setenv("PATH", originalPathEnv)

		if !cmp.Equal(tt.preinstalledTools, []preinstalledTool{}) {
			err = os.RemoveAll(preinstallTempDir)
			if err != nil {
				t.Fatal(err)
			}
		}

		err = os.RemoveAll(installTempDir)
		if err != nil {
			t.Fatal(err)
		}

		apiServer.Close()
		downloadServer.Close()
	}
}

func setupInstallTempDir(t *testing.T, tt test) (installTempDir string) {
	installTempDir, err := ioutil.TempDir("", "toolctl-test-install-*")
	if err != nil {
		t.Fatal(err)
	}
	if !tt.installDirNotInPath {
		err = os.Setenv("PATH", os.ExpandEnv(installTempDir+":$PATH"))
		if err != nil {
			t.Fatal(err)
		}
	}
	return
}
