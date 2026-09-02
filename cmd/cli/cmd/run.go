/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [file1.toml folder/ file2.toml]",
	Aliases: []string{"r"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias r, Run collection or request",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: complete run for cli
		fmt.Println("Run Command called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
