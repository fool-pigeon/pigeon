// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

	"github.com/fool-pigeon/pigeon/core"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/pgrest"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type PrintConfig struct {
	onlyPass bool
	onlyFail bool
	noTests  bool
	showRaw  bool
	rawLine  int
}

var printConfig PrintConfig

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&printConfig.onlyPass, "only-pass", false, "only show passed result of requet")
	runCmd.Flags().BoolVar(&printConfig.onlyFail, "only-fail", false, "only show failed result of requet")
	runCmd.Flags().BoolVar(&printConfig.noTests, "no-tests", false, "not show the results of tests")
	runCmd.Flags().BoolVar(&printConfig.showRaw, "show-raw", false, "show raw response")
	runCmd.Flags().IntVarP(&printConfig.rawLine, "line", "l", -1, "Set the number of lines displayed for raw")

	runCmd.Flags().StringP("run", "r", "", "set run file name in collectinon,e.g. dev means .pg/run/dev.toml")
	runCmd.Flags().String("run-file", "", "set full run file path,like .pg/run/dev.toml")
	runCmd.Flags().StringP("conf", "c", fsx.DefaultRootConfFile, "set root conf file name in collection,such as dev mean that .pg/conf/dev")
	runCmd.Flags().String("conf-file", "", "set full conf file path,like .pg/conf/dev")
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [paths]",
	Example: "pigeon run file1.toml folder/ file2.toml",
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
		tagDir := filepath.Join(root, fsx.RootTag)

		var runFilePath string
		if name, _ := cmd.Flags().GetString("run"); name != "" {
			name = strings.TrimSuffix(name, ".toml") + ".toml"
			runFilePath = filepath.Join(tagDir, fsx.RootRun, name)
		}
		if p, _ := cmd.Flags().GetString("run-file"); p != "" {
			runFilePath = p
		}

		var runPaths core.RunPaths
		if err := fsx.ReadToml(runFilePath, &runPaths); err != nil {
			return err
		}
		// start path
		if runPaths.Start == "" {
			if runPaths.Start, err = filepath.Abs("."); err != nil {
				return err
			}
		}
		// check path valid
		runPaths = runPaths.ToAbs(root)
		if err := runPaths.Check(); err != nil {
			return err
		}
		// tasks
		if runPaths.Tasks == nil {
			runPaths.Tasks = make([]string, 0)
		}
		if len(args) > 0 {
			for _, path := range args {
				runPaths.Tasks = append(runPaths.Tasks, filepath.Join(runPaths.Start, path))
			}
		} else if len(runPaths.Tasks) <= 0 {
			runPaths.Tasks = append(runPaths.Tasks, runPaths.Start)
		}

		// root conf path
		var rootConfFilePath string
		if name, _ := cmd.Flags().GetString("conf"); name != "" {
			name = strings.TrimSuffix(name, ".toml") + ".toml"
			rootConfFilePath = filepath.Join(tagDir, fsx.RootConf, name)
		}
		if p, _ := cmd.Flags().GetString("conf-file"); p != "" {
			rootConfFilePath = p
		}

		var rootConf core.RootConf
		if err := fsx.ReadToml(rootConfFilePath, &rootConf); err != nil {
			pterm.Warning.Printfln("not found %s,use empty conf", rootConfFilePath)
			rootConf = core.RootConf{}
			return err
		}

		// Ctrl+C to cancel
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		taskTotal, taskPassed := 0, 0
		rp := core.RunInput{
			RootConf: rootConf,
			Start:    runPaths.Start,
			Tasks:    runPaths.Tasks,
			TraceHandler: func(t *protocol.Trace) {
				pterm.Info.Println(t)
			},
			ResultHandler: func(res core.RunResult) {
				taskResPrint(res)
				taskTotal++
				if res.Err == nil && res.Res.Err == nil && res.Res.Success {
					taskPassed++
				}
			},
		}

		if err := core.Run(ctx, rp); err != nil {
			return err
		}

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

func taskResPrint(res core.RunResult) {

	// filter result
	isPassed := res.Err == nil && res.Res.Err == nil && res.Res.Success
	if printConfig.onlyFail && isPassed {
		return
	}
	if printConfig.onlyPass && !isPassed {
		return
	}

	printFailed := func(a ...any) { pterm.Println(pterm.LightRed(a...)) }
	printPassed := func(a ...any) { pterm.Println(pterm.LightGreen(a...)) }
	printErr := func(a ...any) { pterm.Println(pterm.Red(a...)) }
	// printSuccess := func(a ...any) { pterm.Println(pterm.Green(a...)) }
	printInfo := func(a ...any) { pterm.Println(pterm.LightBlue(a...)) }

	if res.Err != nil {
		pterm.Error.Println(res.Path, res.Err)
		return
	}
	rRes := res.Res
	req := rRes.Request
	if rRes.Err != nil {
		pterm.Error.Println(res.Path, req.ID, rRes.Err)
		return
	}

	// print status
	var sb strings.Builder
	sb.WriteString(res.Path)
	sb.WriteRune(' ')
	sb.WriteString(string(req.Protocol))
	sb.WriteRune(' ')
	sb.WriteString(rRes.Status)
	sb.WriteRune(' ')
	sb.WriteString(rRes.Duration.Round(time.Millisecond / 10).String())
	if rRes.Success {
		pterm.Success.Println(sb.String())
	} else {
		pterm.Error.Println(sb.String())
	}

	if !printConfig.noTests {
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
	}

	if printConfig.showRaw {
		var out string
		switch raw := rRes.Raw.(type) {
		case *pgrest.Response:
			var buf bytes.Buffer
			if err := json.Indent(&buf, raw.Bytes(), "", "  "); err != nil {
				out = raw.String()
			} else {
				out = buf.String()
			}
			// TODO more protocol out printer
		}

		if rawLine := printConfig.rawLine; rawLine > 0 {
			printInfo(strings.Join(strings.SplitN(out, "\n", rawLine+1)[:rawLine], "\n"))
		} else {
			printInfo(out)
		}
	}
}
