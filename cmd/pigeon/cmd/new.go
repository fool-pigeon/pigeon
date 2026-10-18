// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/protocol/httpx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:     "new [files]",
	Aliases: []string{"n"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias n, Create request into new .toml file",
	Example: `pg new r1 r2 -p http`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    createRequest,
}

func init() {
	newCmd.Flags().StringP("protocol", "p", "", "the protocol of request: http,grpc,ws,gql")
	rootCmd.AddCommand(newCmd)
}

func createRequest(cmd *cobra.Command, args []string) error {
	var protocol string

	// TODO cmd/cli: pg new, more request type
	switch p, _ := cmd.Flags().GetString("protocol"); p {
	default:
		return fmt.Errorf("not impl %q protocol yet.", protocol)
	case request.ProtocolREST, "":
		protocol = request.ProtocolREST
	}

	errs := make([]error, 0, len(args))
	for _, name := range args {
		rName, _ := strings.CutSuffix(name, ".toml")
		fName := rName + ".toml"
		// skip if exist or special error
		isExist, err := fsx.IsFileExist(fName)
		if isExist {
			errs = append(errs, fmt.Errorf("%q %w", fName, fsx.ErrExist))
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%q %w", fName, err))
			continue
		}

		// skip if error
		switch protocol {
		case request.ProtocolREST:
			if err := fsx.WriteToml(fName, httpx.NewReqData(rName).SetMethod(httpx.MethodGet)); err != nil {
				errs = append(errs, fmt.Errorf("%w", err))
				continue
			}
		}

		fmt.Printf("Created: %v\n", fName)
	}

	return errors.Join(errs...)
}
