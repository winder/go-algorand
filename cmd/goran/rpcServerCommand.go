// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/algorand/go-algorand/cmd/goran/internal/rpc"
)

func init() {
	goranRoot.AddCommand(rpcServer)
}

var rpcServer = makeRPCServerCommand()

type rpcServerConfig struct {
	port   uint64
	binDir string
	tmpDir string
	clean  bool
}

func makeRPCServerCommand() *cobra.Command {
	var config rpcServerConfig

	cmd := &cobra.Command{
		Use:   "rpc-server",
		Short: "Serve a JSON-RPC interface for automation utilities.",
		Run: func(cmd *cobra.Command, args []string) {
			if config.clean {
				err := os.RemoveAll(config.tmpDir)
				if err != nil {
					panic(err)
				}
			}
			if err := os.MkdirAll(config.tmpDir, 0755); err != nil {
				panic(err)
			}
			server, err := rpc.MakeServer(config.binDir, config.tmpDir)
			if err != nil {
				panic(err)
			}

			http.Handle("/rpc", server)
			http.ListenAndServe(fmt.Sprintf(":%d", config.port), nil)
		},
	}

	// TODO: Should this use the ":<port>" pattern?
	cmd.Flags().Uint64VarP(&config.port, "port", "p", 60544, "Port to server the RPC server.")
	cmd.Flags().BoolVar(&config.clean, "clean", true, "When set the temp dir is deleted on start.")
	cmd.Flags().StringVarP(&config.binDir, "binDir", "b", "", "Directory containing goal.")
	cmd.Flags().StringVarP(&config.tmpDir, "tmpDir", "t", "", "Temporary directory scratch space.")
	//cmd.MarkFlagRequired("binDir")
	cmd.MarkFlagRequired("tmpDir")

	return cmd
}
