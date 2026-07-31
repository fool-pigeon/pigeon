/*
Copyright © 2026 homoarea homoarea114@gmail.com
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
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("new called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
