/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias e, Manage the env vars",
	Long: `The env vars will be stored in the dotenv file named .env
Env vars will be loaded from the root .pg/ directory to current directory and will be overwritten by new ones.
`,
}
var (
	initEnvCmd = &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "alias i, Add or reset .env file in current directory",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, _ := cmd.Flags().GetStringSlice("tags")
			if len(tags) == 0 {
				tags = []string{""}
			}
			for _, tag := range tags {
				fName := env.FileName(tag)
				if err := fsx.WriteFile(fName, []byte{}); err != nil {
					return err
				}
				fmt.Println("Created:", fName)
			}
			return nil
		},
	}
	listEnvCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"l"},
		Short:   "alias ls, List env vars",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			envs, err := env.LoadFromPath(path)
			if err != nil {
				return err
			}
			envs.Print()
			return nil
		},
	}
)

func init() {
	envCmd.AddCommand(initEnvCmd, listEnvCmd)
	rootCmd.AddCommand(envCmd)

	initEnvCmd.Flags().StringSliceP("tags", "t", nil, "init the env file named .env-tag")

	listEnvCmd.Flags().StringP("tag", "t", "", "switch the config file named .env-tag")
}
