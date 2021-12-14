package api_test

import (
	"testing"

	"github.com/toolctl/toolctl/internal/api"
)

func TestNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    api.NotFoundError
		want string
	}{
		{
			name: "default",
			e:    api.NotFoundError{},
			want: "could not be found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := api.NotFoundError{}
			if got := e.Error(); got != tt.want {
				t.Errorf("NotFoundError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
