package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

// RequireConfigString returns the value of the specified key, or an error if
// the key is not found.
func RequireConfigString(key string) (value string, err error) {
	value = viper.GetString(key)
	if value == "" {
		err = fmt.Errorf("config key '%s' could not be found", key)
	}
	return
}
