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

package rpc

import (
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"

	"github.com/algorand/go-algorand/cmd/goran/internal/core"
)

// makeWatchdogCallback generates a callback function that either 1. does
// nothing if idleTimeout is 0, or 2. kicks a watchdog timer that will send a kill
// signal when it expires.
func makeWatchdogCallback(kill chan os.Signal, idleTimeout time.Duration) func() {
	// If Timeout is nil, then we will not kill kmd after a timeout
	if idleTimeout == 0 {
		return func() {}
	}

	timer := time.AfterFunc(idleTimeout, func() {
		fmt.Println("killing rpc server after timeout.")
		kill <- syscall.SIGINT
	})

	// Return a callback function that resets the timer if it hasn't
	// already expired
	return func() {
		// From docs: For a timer created with AfterFunc(d, f), if
		// t.Stop returns false, then the timer has already expired
		// and the function f has been started in its own goroutine
		if timer.Stop() {
			timer.Reset(idleTimeout)
		}
	}
}

// MakeServer creates the RPC server handler.
func MakeServer(binDir, tmpDir string, kill chan os.Signal, idleTimeout time.Duration) (http.Handler, error) {
	idleRefreshCallback := makeWatchdogCallback(kill, idleTimeout)

	// TODO: if binDir is empty, look for goal/algod on path.
	nw, err := core.MakeNetwork(binDir, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("MakeServer() failed to make core: %w", err)
	}

	coreNet := &Network{
		core: nw,
	}

	coreSystem := &System{
		core: &core.System{},
	}

	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	s.RegisterService(coreNet, "")
	s.RegisterService(coreSystem, "")
	s.RegisterBeforeFunc(func(_ *rpc.RequestInfo) {
		idleRefreshCallback()
	})

	return s, nil
}
