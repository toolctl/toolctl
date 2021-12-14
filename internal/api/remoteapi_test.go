package api

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func Test_remoteAPI_GetLocalAPIFS(t *testing.T) {
	testFS := afero.NewMemMapFs()

	type fields struct {
		BaseURL    *url.URL
		LocalAPIFS afero.Fs
	}
	tests := []struct {
		name   string
		fields fields
		wantFs afero.Fs
	}{
		{
			name: "should work",
			fields: fields{
				BaseURL:    &url.URL{},
				LocalAPIFS: testFS,
			},
			wantFs: testFS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := remoteAPI{
				BaseURL:    tt.fields.BaseURL,
				LocalAPIFS: tt.fields.LocalAPIFS,
			}
			if gotFs := a.GetLocalAPIFS(); !reflect.DeepEqual(gotFs, tt.wantFs) {
				t.Errorf("remoteAPI.GetLocalAPIFS() = %v, want %v", gotFs, tt.wantFs)
			}
		})
	}
}

func Test_remoteAPI_GetLocation(t *testing.T) {
	type fields struct {
		BaseURL    *url.URL
		LocalAPIFS afero.Fs
	}
	tests := []struct {
		name   string
		fields fields
		want   Location
	}{
		{
			name: "should work",
			fields: fields{
				BaseURL:    &url.URL{},
				LocalAPIFS: nil,
			},
			want: Remote,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := remoteAPI{
				BaseURL:    tt.fields.BaseURL,
				LocalAPIFS: tt.fields.LocalAPIFS,
			}
			if got := a.GetLocation(); got != tt.want {
				t.Errorf("remoteAPI.GetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
