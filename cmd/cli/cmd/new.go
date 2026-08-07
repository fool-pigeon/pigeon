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
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias n, Create a request into a new .toml file",
	Example: `pg new -n example -p http`,
	Args:    cobra.RangeArgs(0, 1),
	RunE:    createRequest,
}

var meta request.RequestMeta

func init() {
	newCmd.Flags().StringVarP(&meta.Name, "name", "n", "", "name of the file without suffix")
	newCmd.Flags().StringVarP(&meta.Protocol, "protocol", "p", "", "the protocol of request: http,grpc,ws,gql")
	rootCmd.AddCommand(newCmd)
}

func createRequest(cmd *cobra.Command, args []string) error {
	if meta.Name == "" {
		if err := form.RequestName(&meta.Name); err != nil {
			return err
		}
	}
	if meta.Protocol == "" {
		if err := form.ReqeustProtocol(&meta.Protocol); err != nil {
			return err
		}
	}

	var r request.Request
	switch meta.Protocol {
	case request.P_HTTP:
		if err := fillHttpReq(&r); err != nil {
			return err
		}
	default:
		return errors.New("not impl this protocol yet.")
	}

	return fsx.WriteToml(meta.Name+".toml", r)
}

func fillHttpReq(req *request.Request) error {
	r := httpx.Request{RequestMeta: meta}
	if err := form.HttpForm(&r); err != nil {
		return err
	}
	*req = r
	return nil
}
