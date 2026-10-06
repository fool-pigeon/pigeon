// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	GroupID: GROUP_ID_MANAGE,
	Short:   "alias e, Manage the env vars",
	Long: `The env vars will be stored in the dotenv file and in .pg/env
Env vars will be overwritten by new ones.
`,
}
var (
	listEnvCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"l"},
		Short:   "alias ls, List env file names or vars",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			envDir, err := fsx.EnvDir(path)
			if err != nil {
				return err
			}
			// list vars if need
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				envFile := filepath.Join(envDir, name)
				info, err := os.Lstat(envFile)
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return fmt.Errorf("%s:%w", envFile, fsx.ErrNotFile)
				}
				e, err := env.LoadFromFile(envFile)
				if err != nil {
					return err
				}
				pterm.Info.Println(envFile)
				e.Print()
				return nil
			}
			entries, err := os.ReadDir(envDir)
			if err != nil {
				return err
			}
			pterm.Info.Println("env files in this collection:")
			for _, entry := range entries {
				if !entry.IsDir() {
					fmt.Println(entry.Name())
				}
			}
			return nil
		},
	}
)

func init() {
	envCmd.AddCommand(listEnvCmd)
	rootCmd.AddCommand(envCmd)

	listEnvCmd.Flags().StringP("name", "n", "", "list env vars in the .pg/env/<name>")
}
