/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [file1.toml folder/ file2.toml]",
	Aliases: []string{"r"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias r, Run collection or request",
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: complete run for cli
		paths := make([]string, 0)
		if len(args) > 0 {
			paths = append(paths, args...)
		} else {
			paths = append(paths, ".")
		}

		tag, _ := cmd.Flags().GetString("tag")
		showRaw, _ := cmd.Flags().GetBool("raw")
		core.Run(paths, func(res request.Result) {
			if res.Err != nil {
				pterm.Error.Println(res.ReqName, res.Err)
				return
			}
			if showRaw {
				pterm.Success.Println(res.ReqName, res.Raw)
			}
		}, tag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("tag", "", "run with config and env which has tag")
	runCmd.Flags().Bool("raw", false, "show raw response")
}
