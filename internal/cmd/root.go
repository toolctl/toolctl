package cmd

import (
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	versionFlag bool
)

// NewRootCmd returns the root command.
func NewRootCmd(toolctlWriter io.Writer, localAPIFS afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "toolctl",
		Short: "toolctl controls your tools",
		Example: `  # Get information about installed tools
  toolctl info

  # Install tools
  toolctl install minikube k9s

  # Upgrade supported tools
  toolctl upgrade`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				return printVersion(toolctlWriter)
			}

			return cmd.Help()
		},
	}

	// Flags
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "display the version of toolctl")

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path of the config file (default is $HOME/.config/toolctl/config.yaml)")

	// Hidden persistent flags
	rootCmd.PersistentFlags().Bool("local", false, "Use the local API")
	err := rootCmd.PersistentFlags().MarkHidden("local")
	if err != nil {
		panic(err)
	}

	// Commands
	rootCmd.AddCommand(newInfoCmd(toolctlWriter, localAPIFS))
	rootCmd.AddCommand(newInstallCmd(toolctlWriter, localAPIFS))
	rootCmd.AddCommand(newListCmd(toolctlWriter, localAPIFS))
	rootCmd.AddCommand(newUpgradeCmd(toolctlWriter, localAPIFS))
	rootCmd.AddCommand(newVersionCmd(toolctlWriter))

	// Hidden commands
	rootCmd.AddCommand(newAPICmd(toolctlWriter, localAPIFS))

	return rootCmd
}
