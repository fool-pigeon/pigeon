/*
Copyright © 2026 homoarea homoarea114@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"slices"

	"charm.land/huh/v2"
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

var (
	meta request.RequestMeta
)

func init() {
	newCmd.Flags().StringVarP(&meta.Name, "name", "n", "", "name of the file without suffix")
	newCmd.Flags().StringVarP(&meta.Protocol, "protocol", "p", "", "the protocol of request: http,grpc,ws,gql")
	rootCmd.AddCommand(newCmd)
}

func createRequest(cmd *cobra.Command, args []string) error {
	if meta.Name == "" {
		form := formWrap(huh.NewGroup(
			huh.NewInput().
				Title("Input the name of request:").
				Validate(func(s string) error {
					if s == "" {
						return errors.New("the name is empty.")
					}
					return nil
				}).
				Value(&meta.Name),
		))
		if err := form.Run(); err != nil {
			return err
		}
	}
	if meta.Protocol == "" || !slices.Contains(request.Protocols, meta.Protocol) {
		form := formWrap(
			huh.NewGroup(
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

	var r request.Request
	switch meta.Protocol {
	case request.P_HTTP:
		if err := createHttpReq(&r); err != nil {
			return err
		}
	default:
		return errors.New("not impl this protocol yet.")
	}
	return nil
}
func createHttpReq(req *request.Request) error {
	// TODO: use the form to create a http request
	var r httpx.Request
	const (
		shouldSetPathParams  = "PathParams"
		shouldSetQueryParams = "QueryParams"
		shouldSetHeaders     = "Headers"
		shouldSetBody        = "Body"
		shouldSetSettings    = "Settings"
		shouldSetDesc        = "Description"
	)

	// var nextActions []string
	basicInfoForm := formWrap(
		huh.NewGroup(
			huh.NewInput().
				Title("Input URL:").
				Validate(func(s string) error {
					return r.ParseRawURL(s)
				}),
			huh.NewSelect[string]().
				Title("Select Method").
				Options(huh.NewOptions(httpx.Methods...)...).
				Value(&r.Method),
		),
	)
	if err := basicInfoForm.Run(); err != nil {
		return err
	}
	fmt.Println(r)

	return nil
}
