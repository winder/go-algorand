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
