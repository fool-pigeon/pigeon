/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias i, Init the project struct",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		// create root tag dir
		tagDir := filepath.Join(currentDir, fsx.RootTagDirName)
		if err := fsx.Mkdir(tagDir); err != nil {
			return err
		}
		fmt.Println("Create tag dir:", tagDir)

		// create config file
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if len(tags) == 0 {
			tags = []string{""}
		}
		for _, tag := range tags {
			fName := filepath.Join(tagDir, config.FileName(tag))
			if err := fsx.WriteToml(fName, config.New()); err != nil {
				return err
			}
			fmt.Println("Created:", fName)

			fName = filepath.Join(tagDir, env.FileName(tag))
			if err := fsx.WriteFile(fName, []byte{}); err != nil {
				return err
			}
			fmt.Println("Created:", fName)

		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringSliceP("tags", "t", nil, "init the config file named config-tag.toml")
}
