/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/fsx"
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
		// TODO: generate report of run results such as html,json,toml

		root, err := fsx.RootDir(".")
		if err != nil {
			return err
		}

		paths := make([]string, 0)
		if len(args) > 0 {
			paths = append(paths, args...)
		} else {
			paths = append(paths, ".")
		}

		showRaw, _ := cmd.Flags().GetBool("raw")
		rawLine, _ := cmd.Flags().GetInt("line")

		envName, _ := cmd.Flags().GetString("env")
		envFile, _ := cmd.Flags().GetString("env-file")
		if envFile == "" {
			envFile = filepath.Join(root, envName)
		}
		// Ctrl+C to cancel
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		rp := core.RunPaths{
			Root:   root,
			Tasks:  paths,
			Config: filepath.Join(root, fsx.ConfigDefaultName),
			Env:    envFile,
		}

		taskTotal, taskPassed := 0, 0
		core.Run(ctx, rp, func(res core.RunResult) {
			taskResPrint(res, showRaw, rawLine)
			taskTotal++
			if res.Err == nil && res.ReqRes.Err == nil && res.ReqRes.Success {
				taskPassed++
			}
		})

		// stats the results of run
		status := pterm.LightGreen("✓ PASS")
		stats := pterm.Sprintf("%v (%v Passed, %v Failed)", taskTotal, pterm.LightGreen(taskPassed), (taskTotal - taskPassed))
		if taskPassed != taskTotal {
			status = pterm.LightRed("✗ FAIL")
			stats = pterm.Sprintf("%v (%v Passed, %v Failed)", taskTotal, pterm.LightGreen(taskPassed), pterm.LightRed(taskTotal-taskPassed))
		}
		tableData := pterm.TableData{
			{"Metric", "Value"},
			{"Status", status},
			{"Requests", stats},
		}
		pterm.Println()
		pterm.Info.Println("Stats:")
		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("env-file", "", "set full env file path")
	runCmd.Flags().StringP("env", "e", "", "set env file name in collection,such as dev mean that .pg/envs/dev")
	runCmd.Flags().BoolP("raw", "r", false, "show raw response")
	runCmd.Flags().IntP("line", "l", -1, "Set the number of lines displayed for raw")
}
func taskResPrint(res core.RunResult, showRaw bool, rawLine int) {

	printFailed := func(a ...any) { pterm.Println(pterm.LightRed(a...)) }
	printPassed := func(a ...any) { pterm.Println(pterm.LightGreen(a...)) }
	printErr := func(a ...any) { pterm.Println(pterm.Red(a...)) }
	// printSuccess := func(a ...any) { pterm.Println(pterm.Green(a...)) }
	printInfo := func(a ...any) { pterm.Println(pterm.LightBlue(a...)) }

	if res.Err != nil {
		pterm.Error.Println(res.Path, res.Err)
		return
	}

	// for test view
	// printInfo(pterm.Sprintf("%+v", res))

	rRes := res.ReqRes
	if rRes.Err != nil {
		pterm.Error.Println(rRes.ReqName, rRes.Err)
		return
	}

	// print status
	if rRes.Success {
		pterm.Success.Println(rRes.ReqName, rRes.Status, rRes.Duration.Round(time.Millisecond/10))
	} else {
		pterm.Error.Println(rRes.ReqName, rRes.Status, rRes.Duration.Round(time.Millisecond/10))
	}

	for _, t := range rRes.Tests {
		if t.Err != nil {
			printErr("test: ", t.Name, t.Err)
			continue
		}
		if t.Passed {
			printPassed("    ✓: ", t.Name)
		} else {
			printFailed("    ✗: ", t.Name)
		}
	}
	if showRaw {
		var out string
		switch raw := rRes.Raw.(type) {
		case *httpx.Response:
			var buf bytes.Buffer
			if err := json.Indent(&buf, raw.Bytes(), "", "  "); err != nil {
				out = raw.String()
			} else {
				out = buf.String()
			}
			// TODO more protocol out printer
		}
		if rawLine > 0 {
			printInfo(strings.Join(strings.SplitN(out, "\n", rawLine+1)[:rawLine], "\n"))
		} else {
			printInfo(out)
		}
	}
}
