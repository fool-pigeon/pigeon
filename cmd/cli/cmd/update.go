/*
Copyright © 2026 homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"errors"

	"github.com/homoarea/pigeon/internal/form"
	"github.com/homoarea/pigeon/internal/fsx"
	"github.com/homoarea/pigeon/internal/request"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update [file_path]",
	Aliases: []string{"u"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias u, Update a request's config",
	Args:    cobra.ExactArgs(1),
	RunE:    updateRequest,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateRequest(cmd *cobra.Command, args []string) error {
	filepath := args[0]

	var meta request.RequestMeta
	if err := fsx.ReadToml(filepath, &meta); err != nil {
		return err
	}

	var req request.Request
	switch meta.Protocol {
	case request.P_HTTP:
		var r httpx.Request
		if err := fsx.ReadToml(filepath, &r); err != nil {
			return err
		}
		if err := form.HttpForm(&r); err != nil {
			return err
		}

		req = r

	default:
		return errors.New(pterm.Sprintf("not impl %q protocol yet.", meta.Protocol))
	}

	return fsx.WriteToml(filepath, &req)
}
