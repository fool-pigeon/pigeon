/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/homoarea/pigeon/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias i, Create a .toml file named .pg as root tag",
	Example: "pg init\n#or\npg init <dir_name>",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := filepath.Abs(".")
		if err != nil {
			return err
		}
		if len(args) > 0 {
			name = filepath.Join(name, args[0])
			if err := os.MkdirAll(name, os.FileMode(0755)); err != nil {
				return err
			}
		}
		name = filepath.Join(name, config.RootTagName)
		f, err := os.Create(name)
		if err != nil {
			return err
		}
		fmt.Println("created: ", f.Name())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
