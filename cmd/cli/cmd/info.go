/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// infoCmd represents the inspect command
var infoCmd = &cobra.Command{
	Use:     "info  <file_name>",
	Aliases: []string{"i"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias i, Inspect the request config",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		printField := func(title string, isAddLine bool, content ...string) {
			var sb strings.Builder
			sb.WriteString(pterm.Cyan(title, ": "))
			if isAddLine {
				sb.WriteString("\n")
			}
			for _, v := range content {
				for _, line := range strings.Split(v, "\n") {
					if isAddLine {
						sb.WriteString("  ")
					}
					sb.WriteString(line)
					sb.WriteString("\n")
				}
			}
			pterm.Print(sb.String())
		}

		formatValues := func(data map[string][]string, sep ...string) []string {
			s := ", "
			if len(sep) > 0 {
				s = sep[0]
			}
			values := make([]string, 0, len(data))
			for k, v := range data {
				values = append(values, pterm.Sprintf("%v %v", pterm.LightBlue(k, ":"), strings.Join(v, pterm.LightBlue(s))))
			}
			return values
		}

		flattenValues := func(data map[string][]string) []string {
			values := make([]string, 0, len(data))
			for k, v := range data {
				for _, vv := range v {
					values = append(values, pterm.Sprintf("%v %v", pterm.LightBlue(k, ":"), vv))
				}
			}
			return values
		}

		var meta request.ReqMeta
		b, err := fsx.ReadFile(args[0])
		if err != nil {
			return err
		}
		if err := toml.Unmarshal(b, &meta); err != nil {
			return err
		}

		printField("Name", false, meta.ReqName())
		printField("Protocol", false, meta.ReqProtocol())
		switch meta.ReqProtocol() {
		default:
			return fmt.Errorf("the protocol %q not implement yet.", meta.Protocol)
			// TODO cmd/cli: pg info, more request type
		case request.P_HTTP:
			var r httpx.ReqData
			if err := toml.Unmarshal(b, &r); err != nil {
				return err
			}
			printField("URL", false, r.RawURL())
			printField("Method", false, r.Method)
			if r.Headers != nil {
				printField("Headers", true, formatValues(r.Headers)...)
			}
			switch b := r.Body; {
			case b.Raw != "":
				printField("Body(raw)", true, b.Raw)
			case len(b.Form) > 0:
				printField("Body(form-url-encoded)", true, flattenValues(b.Form)...)
			case len(b.Multipart) > 0:
				fields := make([]string, 0, len(b.Multipart))
				for k, v := range b.Multipart {
					var sb strings.Builder
					if v.Name != "" {
						sb.WriteString(pterm.Sprintln(pterm.LightBlue("Name:"), v.Name))
					}
					if v.Values != nil {
						sb.WriteString(pterm.Sprintln(pterm.LightBlue("Values:"), strings.Join(v.Values, pterm.LightBlue(", "))))
					}
					if v.FileName != "" {
						sb.WriteString(pterm.Sprintln(pterm.LightBlue("FileName:"), v.FileName))
					}
					if v.FilePath != "" {
						sb.WriteString(pterm.Sprintln(pterm.LightBlue("FilePath:"), v.FilePath))
					}
					if v.ContentType != "" {
						sb.WriteString(pterm.Sprintln(pterm.LightBlue("ContentType:"), v.ContentType))
					}
					if k < len(r.Body.Multipart)-1 {
						sb.WriteString("-------------")
					}
					fields = append(fields, sb.String())
				}
				printField("Body(multipart/form-data)", true, fields...)
			case b.Binary != "":
				printField("Body(binary)", true, pterm.Sprint("File: ", b.Binary))
			}

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
