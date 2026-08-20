/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias c, Manage the config",
}
var (
	initConfigCmd = &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "alias i, Create a root tag dir and init the config in current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			currentDir, err := filepath.Abs(".")
			if err != nil {
				return err
			}

			tagDirName := filepath.Join(currentDir, fsx.RootTagDirName)
			if err := fsx.Mkdir(tagDirName); err != nil {
				return err
			}
			fmt.Println("Create tag dir:", tagDirName)

			return nil
		},
	}
	listConfigCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"l"},
		Short:   "alias l, List the config's fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("config list called")
			return nil
		},
	}
)

func init() {
	configCmd.AddCommand(initConfigCmd, listConfigCmd)

	rootCmd.AddCommand(configCmd)
}
