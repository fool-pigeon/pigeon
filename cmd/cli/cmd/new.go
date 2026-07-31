/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:     "new [file_name]",
	Aliases: []string{"n"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias n, Create a request into a new .toml file",
	Example: `pg new example_get -t rest -m get -u localhost`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("new called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
