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
	"net/http"

	"github.com/algorand/go-algorand/cmd/goran/internal/core"
)

// Network wraps the core rpc functions with the signature gorilla looks for.
type Network struct {
	core *core.Network
}

// Start wraps the core Network.Start function.
func (n *Network) Start(r *http.Request, args *core.NetworkStartArgs, reply *core.NetworkStatusReply) error {
	return n.core.Start(args, reply)
}

// Stop wraps the core Network.Stop function.
func (n *Network) Stop(r *http.Request, args *core.NetworkStopArgs, reply *core.NetworkStopReply) error {
	return n.core.Stop(args, reply)
}

// Status wraps the core Network.Status function.
func (n *Network) Status(r *http.Request, args *core.NetworkStatusArgs, reply *core.NetworkStatusReply) error {
	return n.core.Status(args, reply)
}
