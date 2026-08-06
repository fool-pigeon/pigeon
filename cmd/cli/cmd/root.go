/*
Copyright © 2026 homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var ()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "pg",
	Short:   "A cli of API client for API testings created by Go",
	Version: Version,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

const (
	GROUP_ID_MANAGE  string = "management"
	GROUP_ID_REQUEST string = "request"
)

func init() {
	manageGroup := cobra.Group{
		ID:    GROUP_ID_MANAGE,
		Title: "Management Commands",
	}
	requestGroup := cobra.Group{
		ID:    GROUP_ID_REQUEST,
		Title: "Request Commands",
	}

	rootCmd.AddGroup(&manageGroup, &requestGroup)
}
