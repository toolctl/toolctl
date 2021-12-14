package api

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func Test_localAPI_GetLocalAPIFS(t *testing.T) {
	testFS := afero.NewMemMapFs()

	type fields struct {
		BasePath   string
		LocalAPIFS afero.Fs
	}
	tests := []struct {
		name   string
		fields fields
		want   afero.Fs
	}{
		{
			name: "should work",
			fields: fields{
				BasePath:   "",
				LocalAPIFS: testFS,
			},
			want: testFS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := localAPI{
				BasePath:   tt.fields.BasePath,
				LocalAPIFS: tt.fields.LocalAPIFS,
			}
			if got := a.GetLocalAPIFS(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("localAPI.GetLocalAPIFS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localAPI_GetLocation(t *testing.T) {
	type fields struct {
		BasePath   string
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
				BasePath:   "",
				LocalAPIFS: nil,
			},
			want: Local,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := localAPI{
				BasePath:   tt.fields.BasePath,
				LocalAPIFS: tt.fields.LocalAPIFS,
			}
			if got := a.GetLocation(); got != tt.want {
				t.Errorf("localAPI.GetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
