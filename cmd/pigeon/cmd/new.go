// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/sulid"
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
	var protocolInput protocol.ProtocolKind

	// TODO cmd/cli: pg new, more request type
	switch p, _ := cmd.Flags().GetString("protocol"); protocol.ProtocolKind(p) {
	default:
		return fmt.Errorf("not impl %q protocol yet.", p)
	case protocol.ProtocolREST, "":
		protocolInput = protocol.ProtocolREST
	case protocol.ProtocolGRPC, protocol.ProtocolWS, protocol.ProtocolGQL:
		protocolInput = protocol.ProtocolKind(p)
	}

	errs := make([]error, 0, len(args))
	for _, name := range args {
		fName, _ := strings.CutSuffix(name, ".toml")
		fName += ".toml"
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

		id, err := sulid.NewMonotonic()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		req := protocol.Request{}
		req.Id = id
		req.Protocol = protocolInput
		// skip if error
		switch req.Protocol {
		case protocol.ProtocolREST:
			req.Method = protocol.MethodGet
		}

		if err := fsx.WriteToml(fName, req); err != nil {
			errs = append(errs, fmt.Errorf("%s:%w", fName, err))
			continue
		}

		fmt.Printf("Created: %v\n", fName)
	}

	return errors.Join(errs...)
}
