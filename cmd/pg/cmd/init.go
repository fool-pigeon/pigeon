/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias i, Init the project struct",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		if err := core.InitCollection(path); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringSliceP("tags", "t", nil, "init the config file named config-tag.toml")
}
