/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias e, Manage the env vars",
	Long: `The env vars will be stored in the dotenv file named .env
Env vars will be loaded from the root directory to current directory and will be overwritten by new ones.
`,
	RunE: listEnvs,
}
var (
	setCmd = &cobra.Command{
		Use:     "set [var_name] [var_value]",
		Aliases: []string{"s"},
		Short:   "alias s,Add or reset var in the .env",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(args)
			return nil
		},
	}
	removeCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "alias rm,Remove vars from .env",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "alias ls,List env vars",
		RunE:    listEnvs,
	}
)

var global bool

func init() {
	envCmd.AddCommand(setCmd, removeCmd, listCmd)
	rootCmd.AddCommand(envCmd)

	envCmd.PersistentFlags().BoolVarP(&global, "global", "g", false, "manage the gobal env of collection")
}

func listEnvs(cmd *cobra.Command, args []string) error {
	fmt.Println("listEnvs called!")
	return nil
}
