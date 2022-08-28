// Package cmd contains the Cobra CLI.
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// toolctlWriter is a writer that prints to stdout. When testing, we replace
// this with a writer that prints to a buffer.
type toolctlWriter struct{}

func (t toolctlWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// Execute uses the default settings and executes the root command.
func Execute() {
	err := NewRootCmd(toolctlWriter{}, afero.NewOsFs()).Execute()
	if err != nil {
		// Cobra prints the error message
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in the config file if it exists.
func initConfig() {
	home, err := homedir.Dir()
	cobra.CheckErr(err)

	viper.SetDefault("RemoteAPIBaseURL", "https://raw.githubusercontent.com/toolctl/api/main/v0/")
	viper.SetDefault("InstallDir", filepath.Join(home, ".local", "bin"))

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(home + "/.config/toolctl")
		viper.SetConfigName("config")
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error reading config file: %s", err)
		}
	}
}
