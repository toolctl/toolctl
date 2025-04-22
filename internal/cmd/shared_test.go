package cmd_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"io"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/api"
	"github.com/toolctl/toolctl/internal/cmd"
)

// TestArgsToTools tests the ArgsToTools function, ensuring correct parsing of tool names and versions.
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
	name                          string
	notSupportedOnCurrentPlatform bool
	version                       string
	binaryVersion                 string
	downloadURLTemplatePath       string
	ignoredVersions               []string
	onlyOnDownloadServer          bool
	tarGz                         bool
	tarGzSubdir                   string
	tarGzBinaryName               string
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

// setupPreinstallTempDir creates a temporary directory for preinstalled tools and sets up symlinks if needed.
func setupPreinstallTempDir(t *testing.T, tt test) (preinstallTempDir string) {
	preinstallTempDir, err := os.MkdirTemp("", "toolctl-test-install-*")
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

// setupRemoteAPI initializes a mock remote API and download server for testing.
func setupRemoteAPI(supportedTools []supportedTool) (
	toolctlAPI api.ToolctlAPI, apiServer *httptest.Server,
	downloadServer *httptest.Server, err error,
) {
	var downloadServerFS afero.Fs
	downloadServerFS, downloadServer, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS := afero.NewMemMapFs()

	// Create the API content for all supported tools
	var apiFiles []APIFile
	supportedToolNames := make([]string, len(supportedTools))
	for _, supportedTool := range supportedTools {
		var sha256 string
		sha256, err = supportedToolToDownloadFile(downloadServerFS, supportedTool)
		if err != nil {
			return
		}

		if !supportedTool.onlyOnDownloadServer {
			apiFiles = append(
				apiFiles,
				supportedToolToAPIContents(supportedTool, downloadServer.URL, sha256)...,
			)
		}

		supportedToolNames = append(supportedToolNames, supportedTool.name)
	}

	// Create the top-level API content that holds the list of all supported tools
	apiFiles = append(
		apiFiles,
		APIFile{
			Path:     path.Join(localAPIBasePath, "meta.yaml"),
			Contents: "tools:\n  - " + strings.Join(supportedToolNames, "\n  - ") + "\n",
		},
	)

	// Write the API files
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

// setupLocalAPI sets up a mock local API filesystem and optionally creates metadata.
func setupLocalAPI(supportedTools []supportedTool, createTopLevelMeta bool) (
	localAPIFS afero.Fs, downloadServer *httptest.Server, err error,
) {
	var downloadServerFS afero.Fs
	downloadServerFS, downloadServer, err = setupDownloadServer()
	if err != nil {
		return
	}

	localAPIFS = afero.NewMemMapFs()

	// Create the API content for all supported tools
	var apiFiles []APIFile
	supportedToolNames := make([]string, len(supportedTools))
	for _, supportedTool := range supportedTools {
		var sha256 string
		sha256, err = supportedToolToDownloadFile(downloadServerFS, supportedTool)
		if err != nil {
			return
		}

		if !supportedTool.onlyOnDownloadServer {
			apiFiles = append(
				apiFiles,
				supportedToolToAPIContents(supportedTool, downloadServer.URL, sha256)...,
			)
		}

		supportedToolNames = append(supportedToolNames, supportedTool.name)
	}

	// Create the top-level API content that holds the list of all supported tools
	if createTopLevelMeta {
		apiFiles = append(
			apiFiles,
			APIFile{
				Path:     path.Join(localAPIBasePath, "meta.yaml"),
				Contents: "tools:\n  - " + strings.Join(supportedToolNames, "\n  - ") + "\n",
			},
		)
	}

	// Write the API files
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

// setupDownloadServer creates a mock download server using an in-memory filesystem.
func setupDownloadServer() (
	downloadServerFS afero.Fs, downloadServer *httptest.Server, err error,
) {
	downloadServerFS = afero.NewMemMapFs()

	downloadServer = httptest.NewServer(
		http.FileServer(
			afero.NewHttpFs(downloadServerFS).Dir("/"),
		),
	)

	return
}

// calculateSHA256 computes the SHA256 checksum of a tar.gz file for integrity verification.
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

// checkWantOut compares the test output with the expected output or regex.
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

// supportedToolToDownloadFile creates a tar.gz file for a tool and calculates its SHA256 checksum.
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

// createTarGzFile compresses a tar file into a tar.gz file using gzip.
func createTarGzFile(
	tarFilePath string, downloadServerFS afero.Fs,
) (tarGzFilePath string, err error) {
	tarGzFilePath = tarFilePath + ".gz"

	tarFile, err := downloadServerFS.Open(tarFilePath)
	if err != nil {
		return
	}
	defer tarFile.Close()

	tarGzFile, err := downloadServerFS.Create(tarGzFilePath)
	if err != nil {
		return
	}
	defer tarGzFile.Close()

	// Use gzip writer for compression
	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, tarFile)
	if err != nil {
		return
	}

	return
}

// createTarFile creates a tar file from a binary file for testing purposes.
func createTarFile(
	downloadServerFS afero.Fs, filePath string, supportedTool supportedTool,
) (tarFilePath string, err error) {
	tarFilePath = filePath + ".tar"
	tarFile, err := downloadServerFS.Create(tarFilePath)
	if err != nil {
		return
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	inFile, err := downloadServerFS.Open(filePath)
	if err != nil {
		return
	}
	defer inFile.Close()

	fileInfo, err := inFile.Stat()
	if err != nil {
		return
	}

	header := &tar.Header{
		Name: filepath.Join(supportedTool.tarGzSubdir, supportedTool.tarGzBinaryName),
		Size: fileInfo.Size(),
		Mode: int64(fileInfo.Mode()),
	}
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return
	}

	_, err = io.Copy(tarWriter, inFile)
	if err != nil {
		return
	}

	return
}

// createBinaryFile generates a mock binary file for a tool in the test environment.
func createBinaryFile(
	downloadServerFS afero.Fs, supportedTool supportedTool,
) (filePath string, err error) {
	if supportedTool.binaryVersion == "" {
		supportedTool.binaryVersion = supportedTool.version
	}

	filePath = "/" + filepath.Join(
		runtime.GOOS, runtime.GOARCH, supportedTool.version, supportedTool.name,
	)

	err = afero.WriteFile(
		downloadServerFS,
		filePath,
		[]byte(`#!/bin/sh
echo v`+supportedTool.binaryVersion+`
`),
		0644,
	)

	return
}

// supportedToolToAPIContents generates API metadata files for a tool, including download URLs and versions.
func supportedToolToAPIContents(
	supportedTool supportedTool, downloadServerURL string, sha256 string,
) (apiFiles []APIFile) {
	if supportedTool.downloadURLTemplatePath == "" {
		supportedTool.downloadURLTemplatePath = "/{{.OS}}/{{.Arch}}/{{.Version}}/{{.Name}}.tar.gz"
	}

	apiFiles = []APIFile{
		{
			Path: path.Join(localAPIBasePath, supportedTool.name, "meta.yaml"),
			Contents: `description: toolctl test tool
downloadURLTemplate: ` + downloadServerURL + supportedTool.downloadURLTemplatePath + `
ignoredVersions: ['` + strings.Join(supportedTool.ignoredVersions[:], "', '") + `']
homepage: https://toolctl.io/
versionArgs: [version, --short]
`,
		},
	}

	if !supportedTool.notSupportedOnCurrentPlatform {
		apiFiles = append(apiFiles,
			APIFile{
				Path: path.Join(
					localAPIBasePath, supportedTool.name, runtime.GOOS+"-"+runtime.GOARCH,
					"meta.yaml",
				),
				Contents: fmt.Sprintf(`version:
  earliest: %s
  latest: %s
`, supportedTool.version, supportedTool.version),
			},
			APIFile{
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
		)
	}

	return
}

// runInstallUpgradeTests executes tests for install or upgrade commands, verifying results.
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

		installTempDir, err := os.MkdirTemp("", "toolctl-test-install-*")
		if err != nil {
			t.Fatal(err)
		}

		var preinstallTempDir string
		if !cmp.Equal(tt.preinstalledTools, []preinstalledTool{}) {
			preinstallTempDir = setupPreinstallTempDir(t, tt)
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

			command := cmd.NewRootCmd(buf, toolctlAPI.LocalAPIFS())
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
			_ = os.Chmod(preinstallTempDir, 0700)
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

// setupInstallTempDir creates a temporary directory for tool installation and updates PATH if needed.
func setupInstallTempDir(t *testing.T, tt test) (installTempDir string) {
	installTempDir, err := os.MkdirTemp("", "toolctl-test-install-*")
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
