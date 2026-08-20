/*
Copyright © 2026 homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/homoarea/pigeon/internal/fsx"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias c, Init a root config toml file named .pg in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := filepath.Abs(".")
		if err != nil {
			return err
		}
		name = filepath.Join(name, fsx.RootTagDirName)
		f, err := os.Create(name)
		if err != nil {
			return err
		}
		fmt.Println("created: ", f.Name())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
