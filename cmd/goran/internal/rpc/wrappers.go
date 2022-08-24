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
func (n *Network) Start(r *http.Request, args *core.NetworkStartArgs, reply *core.NetworkStartReply) error {
	return n.core.Start(*args, reply)
}

// Destroy wraps the core Network.Destroy function.
func (n *Network) Destroy(r *http.Request, args *core.NetworkDestroyArgs, reply *core.NetworkDestroyReply) error {
	return n.core.Destroy(*args, reply)
}

// Status wraps the core Network.Status function.
func (n *Network) Status(r *http.Request, args *core.NetworkStatusArgs, reply *core.NetworkStatusReply) error {
	return n.core.Status(*args, reply)
}

// List returns a list of all available networks.
func (n *Network) List(r *http.Request, args *core.NetworkListArgs, reply *core.NetworkListReply) error {
	return n.core.List(*args, reply)
}

type System struct {
	core *core.System
}

func (s *System) Shutdown(r *http.Request, args *core.SystemShutdownArgs, reply *core.SystemShutdownReply) error {
	return s.core.Shutdown(*args, reply)
}
