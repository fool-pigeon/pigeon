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
	listConfigCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"l"},
		Short:   "alias l, List the config's fields",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO cli command pg config ls
			suffix, _ := cmd.Flags().GetString("tag")
			fmt.Println(config.FileName(suffix), "list called")
			return nil
		},
	}
)

func init() {
	configCmd.AddCommand(initConfigCmd, listConfigCmd)
	rootCmd.AddCommand(configCmd)

	initConfigCmd.Flags().StringSliceP("tags", "t", nil, "init the config file named config-tag.toml")

	listConfigCmd.Flags().StringP("tag", "t", "", "switch the config file named config-tag.toml")
}
