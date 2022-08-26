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
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/algorand/go-algorand/cmd/goran/internal/rpc"
)

func init() {
	goranRoot.AddCommand(rpcServer)
}

var rpcServer = makeRPCServerCommand()

type rpcServerConfig struct {
	port         uint64
	binDir       string
	tmpDir       string
	clean        bool
	idleShutdown time.Duration
}

func run(config rpcServerConfig) error {
	if config.clean {
		err := os.RemoveAll(config.tmpDir)
		if err != nil {
			panic(err)
		}
	}
	if err := os.MkdirAll(config.tmpDir, 0755); err != nil {
		panic(err)
	}

	// Create a "kill" channel to allow the server to shut down gracefully
	kill := make(chan os.Signal, 1)

	// Timeouts can also send on the kill channel; because signal.Notify
	// will not block, this shouldn't cause an issue. From docs: "Package
	// signal will not block sending to c"
	signal.Notify(kill, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	rcpServer, rpcShutdownFunc, err := rpc.MakeServer(config.binDir, config.tmpDir, kill, config.idleShutdown)
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", config.port),
	}
	http.Handle("/rpc", rcpServer)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("error from ListenAndServe(): %s", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown handling, in case the server times out.
	<-kill
	rpcShutdownFunc()
	return nil
}

func makeRPCServerCommand() *cobra.Command {
	var config rpcServerConfig

	cmd := &cobra.Command{
		Use:   "rpc-server",
		Short: "Serve a JSON-RPC interface for automation utilities.",
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}

	// TODO: Should this use the ":<port>" pattern?
	cmd.Flags().Uint64VarP(&config.port, "port", "p", 60544, "Port to server the RPC server.")
	cmd.Flags().BoolVar(&config.clean, "clean", true, "When set the temp dir is deleted on start.")
	cmd.Flags().StringVarP(&config.binDir, "binDir", "b", "", "Directory containing goal/algod/etc, PATH used if not set.")
	cmd.Flags().StringVarP(&config.tmpDir, "tmpDir", "t", "", "Temporary directory scratch space.")
	//cmd.MarkFlagRequired("binDir")
	cmd.MarkFlagRequired("tmpDir")
	cmd.Flags().DurationVarP(&config.idleShutdown, "idleShutdown", "i", 0, "When set, automatically shutdown RPC server when idle for this duration.")

	return cmd
}
