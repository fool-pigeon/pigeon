/*
Copyright © 2026 homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/homoarea/pigeon/internal/env"
	"github.com/knadh/koanf/v2"
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
		RunE:    setEnv,
	}
	removeCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "alias rm,Remove vars from .env",
		Args:    cobra.ArbitraryArgs,
		RunE:    removeEnvs,
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
	envs, err := env.Load(".")
	if err != nil {
		return err
	}
	envs.Print()
	return nil
}
func setEnv(cmd *cobra.Command, args []string) error {
	k := koanf.New(".")
	if err := env.LoadFromFile(k, env.EnvFileName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	k.Set(args[0], args[1])
	if err := env.Store(k, env.EnvFileName); err != nil {
		return err
	}
	return nil
}
func removeEnvs(cmd *cobra.Command, args []string) error {
	k := koanf.New(".")
	if err := env.LoadFromFile(k, env.EnvFileName); err != nil {
		return err
	}
	for _, envKey := range args {
		k.Delete(envKey)
	}
	if err := env.Store(k, env.EnvFileName); err != nil {
		return err
	}
	return nil
}
