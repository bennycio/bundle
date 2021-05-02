package cli

import (
	"github.com/spf13/cobra"

	_ "embed"
)

const (
	BundleFileName = "bundle.yml"
)

type PluginYML struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

//go:embed bundle.yml
var BundleYml string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Base command for the Bundle CLI",
}

var Force bool

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force the command to run regardless of constraints")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// TODO
}
