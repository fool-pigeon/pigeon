/*
Copyright © 2026 homoarea homoarea114@gmail.com
*/
package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/homoarea/pigeon/internal/request"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:     "new [file_name]",
	Aliases: []string{"n"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias n, Create a request into a new .toml file",
	Example: `pg new example_get`,
	Args:    cobra.RangeArgs(0, 1),
	RunE:    createRequest,
}

var (
	meta request.RequestMeta
)

func init() {
	newCmd.Flags().StringVarP(&meta.Protocol, "protocol", "p", request.P_HTTP, "the protocol of request(default http): http,grpc,ws,gql")
	rootCmd.AddCommand(newCmd)
}

func createRequest(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		meta.Name = args[0]
	} else {
		form := huh.NewForm(
			//
			huh.NewGroup(
				huh.NewInput().
					Title("Input the name of request:").
					Value(&meta.Name),
				huh.NewSelect[string]().
					Title("Pick a protocol.").
					Options(huh.NewOptions(request.Protocols...)...).
					Value(&meta.Protocol),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	}
	fmt.Println(meta)
	return nil
}
func createHttpReq(req *request.Request) error {
	// form := huh.NewGroup()

	return nil
}
