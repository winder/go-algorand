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

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"

	"github.com/algorand/go-algorand/cmd/goran/internal/core"
)

// MakeServer creates the RPC server handler.
func MakeServer(binDir, tmpDir string) (http.Handler, error) {
	// TODO: if binDir is empty, look for goal/algod on path.
	nw, err := core.MakeNetwork(binDir, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("MakeServer() failed to make core: %w", err)
	}
	coreNet := &Network{
		core: nw,
	}

	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	s.RegisterService(coreNet, "")

	return s, nil
}
