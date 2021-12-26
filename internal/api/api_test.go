package api

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	testFS := afero.NewMemMapFs()
	baseURL, _ := url.Parse("http://localhost/")
	basePath := "/tmp/test"

	type args struct {
		localAPIFS afero.Fs
		// cmd is set in the tests
		defaultLocation Location
	}

	tests := []struct {
		name               string
		remoteAPIBaseURL   string
		localAPIBasePath   string
		localFlag          bool
		localFlagUndefined bool
		args               args
		want               ToolctlAPI
		wantErrStr         string
	}{
		{
			name:             "remote",
			remoteAPIBaseURL: baseURL.String(),
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Remote,
			},
			want: &remoteAPI{
				localAPIFS: testFS,
				baseURL:    baseURL,
			},
		},
		{
			name: "remote without RemoteAPIBaseURL",
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Remote,
			},
			wantErrStr: "config key 'RemoteAPIBaseURL' could not be found",
		},
		{
			name:             "local",
			localAPIBasePath: basePath,
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Local,
			},
			want: &localAPI{
				basePath:   basePath,
				localAPIFS: testFS,
			},
		},
		{
			name:             "local through flag",
			localAPIBasePath: basePath,
			localFlag:        true,
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Remote,
			},
			want: &localAPI{
				basePath:   basePath,
				localAPIFS: testFS,
			},
		},
		{
			name: "local without LocalAPIBasePath",
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Local,
			},
			wantErrStr: "config key 'LocalAPIBasePath' could not be found",
		},
		{
			name:               "local flag undefined",
			localFlagUndefined: true,
			args: args{
				localAPIFS:      testFS,
				defaultLocation: Remote,
			},
			wantErrStr: "flag accessed but not defined: local",
		},
	}

	for _, tt := range tests {
		cmd := &cobra.Command{}
		if !tt.localFlagUndefined {
			cmd.Flags().Bool("local", tt.localFlag, "")
		}

		viper.Set("RemoteAPIBaseURL", tt.remoteAPIBaseURL)
		viper.Set("LocalAPIBasePath", tt.localAPIBasePath)

		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.localAPIFS, cmd, tt.args.defaultLocation)
			if (err == nil) != (tt.wantErrStr == "") {
				t.Errorf("New() error = %v, wantErr %v", err, (tt.wantErrStr != ""))
				return
			}
			if err != nil && err.Error() != tt.wantErrStr {
				t.Errorf("New() error = %v, wantErrStr = %v", err, tt.wantErrStr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
