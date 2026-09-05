/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/json"

	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
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
		// TODO: stats the request results
		paths := make([]string, 0)
		if len(args) > 0 {
			paths = append(paths, args...)
		} else {
			paths = append(paths, ".")
		}

		tag, _ := cmd.Flags().GetString("tag")
		showRaw, _ := cmd.Flags().GetBool("raw")

		return core.Run(".", paths, func(res request.Result) {
			printFailed := func(a ...any) { pterm.Println(pterm.LightRed(a...)) }
			printErr := func(a ...any) { pterm.Println(pterm.Red(a...)) }
			printSuccess := func(a ...any) { pterm.Println(pterm.Green(a...)) }
			printInfo := func(a ...any) { pterm.Println(pterm.LightBlue(a...)) }

			if res.Err != nil {
				pterm.Error.Println(res.ReqName, res.Err)
				return
			}
			if res.Success {
				pterm.Success.Println(res.ReqName, res.Status)
			} else {
				pterm.Error.Println(res.ReqName, res.Status)
			}

			for _, tRes := range res.Tests {
				if tRes.Err != nil {
					printErr("test: ", tRes.Name, tRes.Err)
					continue
				}
				if tRes.Passed {
					printSuccess("test success: ", tRes.Name)
				} else {
					printFailed("test failed: ", tRes.Name)
				}
			}
			if showRaw {
				switch raw := res.Raw.(type) {
				case *httpx.Response:
					var out bytes.Buffer
					if err := json.Indent(&out, raw.Bytes(), "", "  "); err != nil {
						printInfo(raw)
						return
					}
					printInfo(out.String())
				}
			}
		}, tag)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("tag", "", "run with config and env which has tag")
	runCmd.Flags().Bool("raw", false, "show raw response")
}
