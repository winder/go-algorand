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
