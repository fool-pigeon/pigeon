/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
