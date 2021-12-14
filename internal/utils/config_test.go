package utils_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/toolctl/toolctl/internal/utils"
)

func TestRequireConfigString(t *testing.T) {
	type args struct {
		key string
	}

	tests := []struct {
		name       string
		args       args
		wantValue  string
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "ok",
			args: args{
				key: "myKey",
			},
			wantValue: "all good",
		},
		{
			name: "missing",
			args: args{
				key: "missing",
			},
			wantValue:  "",
			wantErr:    true,
			wantErrStr: "config key 'missing' could not be found",
		},
	}

	viper.Set("myKey", "all good")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := utils.RequireConfigString(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireConfigString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("RequireConfigString() = %v, want %v", gotValue, tt.wantValue)
			}
			if (err != nil) && (err.Error() != tt.wantErrStr) {
				t.Errorf("RequireConfigString() error = %v, wantErr %v", err, tt.wantErrStr)
			}
		})
	}
}
