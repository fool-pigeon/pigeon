/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// infoCmd represents the inspect command
var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias i, Inspect the request config",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("inspect called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
