// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	"strings"

	"github.com/fool-pigeon/pigeon/cmd/pigeon/cmd/formatter"
	"github.com/fool-pigeon/pigeon/fsx"
	pl "github.com/fool-pigeon/pigeon/protocol"
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

		b, err := fsx.ReadFile(args[0])
		if err != nil {
			return err
		}

		var req pl.Request
		if err := toml.Unmarshal(b, &req); err != nil {
			return err
		}

		printValues := formatter.PrintValues
		formatValues := formatter.FormatValues
		flattenValues := formatter.FlattenValues
		line := formatter.FieldLine

		printValues("Name", false, req.Id.String())
		printValues("Protocol", false, string(req.Protocol))
		switch req.Protocol {
		// TODO cmd/cli: pg info, more request type
		default:
			return fmt.Errorf("the protocol %q not implement yet", req.Protocol)
		case pl.ProtocolREST:
			// TODO: finish http request info
			printValues("URL", false, req.RawURL())
			printValues("Method", false, string(req.Method))
			if req.Header != nil {
				printValues("Headers", true, formatValues(req.Header)...)
			}
			switch b := req.PayloadREST; b.Kind {
			case pl.BodyRaw, pl.BodyJSON, pl.BodyGQL, pl.BodyXML:
				printValues(fmt.Sprintf("Body(%s)", b.Kind), true, b.Raw)
			case pl.BodyBinary:
				printValues("Body(binary)", true, b.BinaryPath)
			case pl.BodyUrlEncoded:
				printValues("Body(url-encoded)", true, flattenValues(b.UrlEncoded)...)
			case pl.BodyMultipart:
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
					if k < len(b.Multipart)-1 {
						sb.WriteString("-------------")
					}
					fields = append(fields, sb.String())
				}
				printValues("Body(multipart/form-data)", true, fields...)
			}

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
