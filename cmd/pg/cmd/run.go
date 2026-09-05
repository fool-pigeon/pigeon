/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

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
		rawLine, _ := cmd.Flags().GetInt("line")

		return core.Run(".", paths, func(res request.Result) {
			printFailed := func(a ...any) { pterm.Println(pterm.LightRed(a...)) }
			printPassed := func(a ...any) { pterm.Println(pterm.LightGreen(a...)) }
			printErr := func(a ...any) { pterm.Println(pterm.Red(a...)) }
			// printSuccess := func(a ...any) { pterm.Println(pterm.Green(a...)) }
			printInfo := func(a ...any) { pterm.Println(pterm.LightBlue(a...)) }

			if res.Err != nil {
				pterm.Error.Println(res.ReqName, res.Err)
				return
			}
			if res.Success {
				pterm.Success.Println(res.ReqName, res.Status, res.Duration.Round(time.Millisecond/10))
			} else {
				pterm.Error.Println(res.ReqName, res.Status, res.Duration.Round(time.Millisecond/10))
			}

			for _, tRes := range res.Tests {
				if tRes.Err != nil {
					printErr("test: ", tRes.Name, tRes.Err)
					continue
				}
				if tRes.Passed {
					printPassed("    ✓: ", tRes.Name)
				} else {
					printFailed("    ✗: ", tRes.Name)
				}
			}
			if showRaw {
				var out string
				switch raw := res.Raw.(type) {
				case *httpx.Response:
					var buf bytes.Buffer
					if err := json.Indent(&buf, raw.Bytes(), "", "  "); err != nil {
						out = raw.String()
					} else {
						out = buf.String()
					}
				}
				if rawLine > 0 {
					printInfo(strings.Join(strings.SplitN(out, "\n", rawLine+1)[:rawLine], "\n"))
				} else {
					printInfo(out)
				}
			}
		}, tag)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("tag", "", "run with config and env which has tag")
	runCmd.Flags().BoolP("raw", "r", false, "show raw response")
	runCmd.Flags().IntP("line", "l", -1, "Set the number of lines displayed for raw")
}
