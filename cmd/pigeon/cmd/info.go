// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	"strings"

	"github.com/fool-pigeon/pigeon/cmd/pigeon/cmd/formatter"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// infoCmd represents the inspect command
var infoCmd = &cobra.Command{
	Use:     "info  [file]",
	GroupID: GROUP_ID_REQUEST,
	Short:   "Inspect the request config",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		var meta request.ReqMeta
		b, err := fsx.ReadFile(args[0])
		if err != nil {
			return err
		}
		if err := toml.Unmarshal(b, &meta); err != nil {
			return err
		}

		printValues := formatter.PrintValues
		formatValues := formatter.FormatValues
		flattenValues := formatter.FlattenValues
		line := formatter.FieldLine

		printValues("Name", false, meta.ReqName())
		printValues("Protocol", false, meta.ReqProtocol())
		switch meta.ReqProtocol() {
		// TODO cmd/cli: pg info, more request type
		default:
			return fmt.Errorf("the protocol %q not implement yet.", meta.Protocol)
		case request.ProtocolREST:
			// TODO: finish http request info
			var r httpx.ReqData
			if err := toml.Unmarshal(b, &r); err != nil {
				return err
			}
			printValues("URL", false, r.RawURL())
			printValues("Method", false, r.Method)
			if r.Header != nil {
				printValues("Headers", true, formatValues(r.Header)...)
			}
			switch b := r.Body; {
			case b.Raw != "":
				printValues("Body(raw)", true, b.Raw)
			case len(b.Form) > 0:
				printValues("Body(form-url-encoded)", true, flattenValues(b.Form)...)
			case len(b.Multipart) > 0:
				fields := make([]string, 0, len(b.Multipart))
				for k, v := range b.Multipart {
					var sb strings.Builder
					sb.WriteString(line("Name", v.Name))
					if v.Value != "" {
						sb.WriteString(line("Values", v.Value))
					}
					if v.FileName != "" {
						sb.WriteString(line("FileName", v.FileName))
					}
					if v.FilePath != "" {
						sb.WriteString(line("FilePath", v.FilePath))
					}
					if v.ContentType != "" {
						sb.WriteString(line("ContentType", v.ContentType))
					}
					if k < len(r.Body.Multipart)-1 {
						sb.WriteString("-------------")
					}
					fields = append(fields, sb.String())
				}
				printValues("Body(multipart/form-data)", true, fields...)
			case b.Binary != "":
				printValues("Body(binary)", true, b.Binary)
			}

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
