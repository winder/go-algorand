package luarunner

import (
	"github.com/algorand/go-algorand/cmd/algomation/core/common"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
)

// Run a lua script.
func Run(p common.Params) error {
	L := lua.NewState()
	luajson.Preload(L)
	registerTxnType(L)
	L.PreloadModule("test", testLoader)
	ncLoader := makeNodeControllerLoader("/Users/shiqi/go/bin", "/Users/shiqi/.algorand/testdir/testnet")
	L.PreloadModule("algod", ncLoader)
	registerNodeControllerType(L)
	L.PreloadModule("algotest", AlgoTestLoader)
	defer L.Close()
	if err := L.DoFile(p.ScriptFile); err != nil {
		return err
	}
	return nil
}
