/*
Copyright © 2026 homoarea <homoarea114@gmail.com>
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

var meta request.RequestMeta

func init() {
	newCmd.Flags().StringVarP(&meta.Name, "name", "n", "", "name of the file without suffix")
	newCmd.Flags().StringVarP(&meta.Protocol, "protocol", "p", "", "the protocol of request: http,grpc,ws,gql")
	rootCmd.AddCommand(newCmd)
}

func createRequest(cmd *cobra.Command, args []string) error {
	fields := []huh.Field{}
	if meta.Name == "" {
		f := huh.NewInput().
			Title("Input the name of request:").
			Validate(func(s string) error {
				if s == "" {
					return errors.New("the name is empty.")
				}
				return nil
			}).
			Value(&meta.Name)
		fields = append(fields, f)
	}
	if meta.Protocol == "" || !slices.Contains(request.Protocols, meta.Protocol) {
		f := huh.NewSelect[string]().
			Title("Pick a protocol.").
			Options(huh.NewOptions(request.Protocols...)...).
			Value(&meta.Protocol)
		fields = append(fields, f)
	}
	if len(fields) > 0 {
		if err := formWrap(huh.NewGroup(fields...)).Run(); err != nil {
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
	const (
		shouldURL = iota
		shouldMethod
		shouldPathParams
		shouldQueryParams
		shouldHeaders
		shouldBody
		shouldSettings
		shouldDesc
	)

	r := httpx.Request{RequestMeta: meta}
	var actions []int
	{
		actionForm := formWrap(
			huh.NewGroup(
				huh.NewMultiSelect[int]().
					Title("Select options you want to set").
					Options(
						huh.Option[int]{Key: "URL", Value: shouldURL}.Selected(true),
						huh.Option[int]{Key: "Method", Value: shouldMethod}.Selected(true),
						huh.Option[int]{Key: "PathParams", Value: shouldPathParams},
						huh.Option[int]{Key: "QueryParams", Value: shouldQueryParams},
						huh.Option[int]{Key: "Headers", Value: shouldHeaders},
						huh.Option[int]{Key: "Body", Value: shouldBody},
						huh.Option[int]{Key: "Settings", Value: shouldSettings},
						huh.Option[int]{Key: "Description", Value: shouldDesc},
					).
					Value(&actions),
			),
		)
		if err := actionForm.Run(); err != nil {
			return err
		}
	}

	// set rawURL,it will auto parse such as query params
	if slices.Contains(actions, shouldURL) {
		f := formWrap(huh.NewGroup(huh.NewInput().Title("Input URL").Validate(func(s string) error { return r.ParseRawURL(s) })))
		if err := f.Run(); err != nil {
			return err
		}
	}

	// set other config
	{
		form := formWrap( // set http method
			huh.NewGroup(huh.NewSelect[string]().Value(&r.Method).Options(huh.NewOptions(httpx.Methods...)...)).
				WithHide(!slices.Contains(actions, shouldMethod)),
			// set path params
			huh.NewGroup(tomlTextFormField(&r.PathParams).Title("PathParams")).
				WithHide(!slices.Contains(actions, shouldPathParams)),
			// set query params
			huh.NewGroup(tomlTextFormField(&r.QueryParams).Title("QueryParams")).
				WithHide(!slices.Contains(actions, shouldPathParams)),
			// set headers
			huh.NewGroup(tomlTextFormField(&r.Headers).Title("Headers")).
				WithHide(!slices.Contains(actions, shouldHeaders)),
			// set settings
			huh.NewGroup(tomlTextFormField(&r.Settings).Title("Settings")).
				WithHide(!slices.Contains(actions, shouldSettings)),
			// set desc
			huh.NewGroup(huh.NewText().Value(&r.Desc).Title("Description")).
				WithHide(!slices.Contains(actions, shouldDesc)),
			// set body
			httpBodyFormGroup(&r.Body).
				WithHide(!slices.Contains(actions, shouldBody)))
		if err := form.Run(); err != nil {
			return err
		}
	}

	// for test
	fmt.Printf("%#v\n", r)

	return nil
}
