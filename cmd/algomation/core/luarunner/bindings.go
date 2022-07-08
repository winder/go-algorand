package luarunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v2"

	gosdk "github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand/nodecontrol"
	luajson "layeh.com/gopher-json"
)

// testLoader basic module test that also has field.
func testLoader(L *lua.LState) int {
	// this mod thing might be a hack.
	var mod *lua.LTable

	start := func(L *lua.LState) int {
		fmt.Println("Start algod here...")
		L.SetField(mod, "state", lua.LString("started"))
		return 0
	}

	var exports = map[string]lua.LGFunction{
		"start": start,
	}

	// register functions to the table
	mod = L.SetFuncs(L.NewTable(), exports)

	// register other stuff
	L.SetField(mod, "state", lua.LString("stopped"))

	// returns the module
	L.Push(mod)
	return 1
}

// makeNodeControllerLoader initializes bindings to node controller with hard coded bin/data dir.
// Example lua:
// 	   local algod = require("algodModule")
// 	   print("Starting node.")
// 	   algod.start()
// 	   print("Getting status, node started.")
// 	   algod.status()
// 	   print("Stopping node.")
// 	   algod.stop()
func makeNodeControllerLoader(bindir, datadir string) lua.LGFunction {
	return func(L *lua.LState) int {
		var mod *lua.LTable
		nc := nodecontrol.MakeNodeController(bindir, datadir)

		var exports = map[string]lua.LGFunction{
			"start": func(L *lua.LState) int {
				nc.StartAlgod(nodecontrol.AlgodStartArgs{})
				return 0
			},
			"stop": func(L *lua.LState) int {
				nc.StopAlgod()
				return 0
			},
			"status": func(L *lua.LState) int {
				c, err := nc.AlgodClient()
				if err != nil {
					fmt.Println("Problem getting client.")
					return 1
				}
				s, err := c.Status()
				if err != nil {
					fmt.Println("Problem getting status.")
					return 1
				}
				fmt.Printf("%v\n", s)
				return 0
			},
		}

		// register functions to the table
		mod = L.SetFuncs(L.NewTable(), exports)

		// returns the module
		L.Push(mod)
		return 1
	}
}

const luaNodeControllerName = "node-controller"

func checkNodeController(L *lua.LState) *nodecontrol.NodeController {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*nodecontrol.NodeController); ok {
		return v
	}
	L.ArgError(1, "node controller expected")
	return nil
}

// registerNodeControllerType initializes bindings to a global node controller type.
// Example lua:
//		local node = algod.new("/home/will/go/bin", "/home/will/nodes/testdir")
//		node:start()
//		node:status()
//		node:stop()
func registerNodeControllerType(L *lua.LState) {
	// Constructor
	newAlgod := func(L *lua.LState) int {
		nc := nodecontrol.MakeNodeController(L.CheckString(1), L.CheckString(2))
		ud := L.NewUserData()
		ud.Value = &nc
		L.SetMetatable(ud, L.GetTypeMetatable(luaNodeControllerName))
		L.Push(ud)
		return 1
	}

	// Type methods
	var methods = map[string]lua.LGFunction{
		"start": func(L *lua.LState) int {
			nc := checkNodeController(L)
			nc.StartAlgod(nodecontrol.AlgodStartArgs{})
			return 1
		},
		"stop": func(L *lua.LState) int {
			nc := checkNodeController(L)
			nc.StopAlgod()
			return 1
		},
		"status": func(L *lua.LState) int {
			nc := checkNodeController(L)
			c, err := nc.AlgodClient()
			if err != nil {
				fmt.Println("Problem getting client.")
				return 1
			}
			s, err := c.Status()
			if err != nil {
				fmt.Println("Problem getting status.")
				return 1
			}
			fmt.Printf("%v\n", s)
			return 1
		},
	}

	// Register new type
	mt := L.NewTypeMetatable(luaNodeControllerName)
	L.SetGlobal("algod", mt)
	L.SetField(mt, "new", L.NewFunction(newAlgod))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}

const luaAlgoTestName = "AlgotTest"

const algoTestTxnType = "txn"
const algoTestAccountType = "account"

func registerTxnType(L *lua.LState) {
	mt := L.NewTypeMetatable(algoTestTxnType)
	L.SetGlobal("txn", mt)
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), txnMethods))
}

var txnMethods = map[string]lua.LGFunction{
	"submit": submit,
}

func submit(L *lua.LState) int {
	ud := L.CheckUserData(1)
	walletID := L.CheckString(2)
	txn := ud.Value.(types.Transaction)
	kmdClient := getKMDClient()
	resp2, err := kmdClient.InitWalletHandle(walletID, "")
	if err != nil {
		fmt.Printf("Error initializing wallet: %s\n", err)
		return 0
	}

	// Extract the wallet handle
	walletHandleToken := resp2.WalletHandleToken

	//extra pk
	resp4, err := kmdClient.ExportKey(walletHandleToken, "", txn.Sender.String())
	if err != nil {
		fmt.Printf("Error extracting secret key: %s\n", err)
		return 0
	}
	privateKey := resp4.PrivateKey

	_, stxn, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return 0
	}

	algodClient := getAlgodClient()
	txid, err := algodClient.SendRawTransaction(stxn).Do(context.Background())
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return 0
	}
	_, err = future.WaitForConfirmation(algodClient, txid, 5, context.Background())
	if err != nil {
		fmt.Printf("wait for confirmation err: %s\n", err)
		return 0
	}

	response, _, err := algodClient.PendingTransactionInformation(txid).Do(context.Background())
	applicationID := response.ApplicationIndex
	L.Push(lua.LNumber(applicationID))
	return 1
}

// AlgoTestLoader defines test methods
// Example lua:
// 	   local t = require("algotest")
//     local addr = t.makeAccount()
// 	   print(addr)
func AlgoTestLoader(L *lua.LState) int {
	var exports = map[string]lua.LGFunction{
		"makeAccount":         makeAccount,
		"createAppFromConfig": createAppFromConfig,
		"createAsa":           createAsa,
		"startPrivateNetwork": startPrivateNetwork,
		"assertAccountStates": assertAccountStates,
	}
	// register functions to the table
	mod := L.SetFuncs(L.NewTable(), exports)

	// returns the module
	L.Push(mod)
	return 1
}

func makeAccount(L *lua.LState) int {
	src := "H6Y3Z3WWVSTI4LNTKFUPVFECG7CRD2PNSVCHY5M35EZ2YGWT66UYTSZ34I"
	kmdClient := getKMDClient()
	resp0, err := kmdClient.ListWallets()
	if err != nil {
		fmt.Printf("error listing wallets: %s\n", err)
		return 0
	}
	fmt.Printf("Got %d wallet(s): %s\n", len(resp0.Wallets), resp0.Wallets[0].ID)
	// Get a wallet handle
	resp2, err := kmdClient.InitWalletHandle(resp0.Wallets[0].ID, "")
	if err != nil {
		fmt.Printf("Error initializing wallet: %s\n", err)
		return 0
	}
	// Extract the wallet handle
	exampleWalletHandleToken := resp2.WalletHandleToken

	resp3, err := kmdClient.GenerateKey(exampleWalletHandleToken)
	if err != nil {
		fmt.Printf("Error generating key: %s\n", err)
		return 0
	}
	fmt.Printf("Generated address %s\n", resp3.Address)

	addr := resp3.Address
	algodClient := getAlgodClient()

	nodeStatus, err := algodClient.Status().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		return 0
	}
	fmt.Printf("algod last round: %d\n", nodeStatus.LastRound)
	//Get the suggested transaction parameters
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return 0
	}
	tx, err := future.MakePaymentTxn(src, addr, 1000000, nil, "", txParams)
	if err != nil {
		fmt.Printf("Error creating transaction: %s\n", err)
		return 0
	}
	// Sign the same transaction with kmd
	fmt.Println("Signing transaction with kmd")
	resp5, err := kmdClient.SignTransaction(exampleWalletHandleToken, "", tx)
	if err != nil {
		fmt.Printf("Failed to sign transaction with kmd: %s\n", err)
		return 0
	}
	L.Push(lua.LString(addr))

	txid, err := algodClient.SendRawTransaction(resp5.SignedTransaction).Do(context.Background())
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return 0
	}

	_, err = future.WaitForConfirmation(algodClient, txid, 5, context.Background())
	if err != nil {
		fmt.Printf("wait for confirmation err: %s\n", err)
		return 0
	}

	return 1
}

//var appID = 1

// Contract a contract type
type Contract struct {
	Contracts map[string]map[string]string
}

var contractConfigs Contract

func createAppFromConfig(L *lua.LState) int {
	creator := L.CheckString(1)
	sender, _ := basics.UnmarshalChecksumAddress(creator)

	contractName := L.CheckString(2)
	// parse contract configs
	filename, _ := filepath.Abs("configs/demo.yml")
	config, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error : %s\n", err)
		return 0
	}
	err = yaml.Unmarshal(config, &contractConfigs)
	if err != nil {
		fmt.Printf("error : %s\n", err)
		return 0
	}
	//fmt.Printf("%+v\n", contractConfigs)
	yaml.Unmarshal(config, &contractConfigs)
	contract1Configs := contractConfigs.Contracts[contractName]
	localint, _ := strconv.ParseInt(contract1Configs["local_int"], 10, 64)
	localbyte, _ := strconv.ParseInt(contract1Configs["local_byte"], 10, 64)
	globalint, _ := strconv.ParseInt(contract1Configs["global_int"], 10, 64)
	globalByte, _ := strconv.ParseInt(contract1Configs["global_byte"], 10, 64)
	extraPages, _ := strconv.ParseInt(contract1Configs["extra_program_pages"], 10, 64)

	globalState := types.StateSchema{
		NumUint:      uint64(globalint),
		NumByteSlice: uint64(globalByte),
	}

	localState := types.StateSchema{
		NumUint:      uint64(localint),
		NumByteSlice: uint64(localbyte),
	}
	// create an app
	algodClient := getAlgodClient()
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return 0
	}

	ops, _ := logic.AssembleStringWithVersion(contract1Configs["approval_program"], 6)
	pd := logic.HashProgram(ops.Program)
	clear, _ := logic.AssembleStringWithVersion(contract1Configs["clear_state_program"], 6)

	txn, err := future.MakeApplicationCreateTxWithExtraPages(true, ops.Program, clear.Program,
		globalState, localState, nil, nil, nil, nil, sp, types.Address(sender), nil, types.Digest{}, [32]byte{}, types.Address{}, uint32(extraPages))
	if err != nil {
		fmt.Printf("error creating an app: %s\n", err)
		return 0
	}

	ud := L.NewUserData()
	ud.Value = txn
	L.Push(ud) // return txn
	L.SetMetatable(ud, L.GetTypeMetatable(algoTestTxnType))

	addr := basics.Address(pd) // return contract address
	L.Push(lua.LString(addr.String()))
	return 2 // return 2 values
}

func createAsa(L *lua.LState) int {
	txn := transactions.Transaction{AssetConfigTxnFields: transactions.AssetConfigTxnFields{
		ConfigAsset: 0,
		AssetParams: basics.AssetParams{
			Total:         0,
			Decimals:      0,
			DefaultFrozen: false,
			UnitName:      "",
			AssetName:     "",
			URL:           "",
			MetadataHash:  [32]byte{},
			Manager:       basics.Address{},
			Reserve:       basics.Address{},
			Freeze:        basics.Address{},
			Clawback:      basics.Address{},
		},
	}}
	ud := L.NewUserData()
	ud.Value = txn
	L.Push(ud)
	L.SetMetatable(ud, L.GetTypeMetatable(luaAlgoTestName))
	return 1
}

func startPrivateNetwork(L *lua.LState) int {
	cmd := exec.Command("./sandbox", "up", "-v")
	cmd.Dir = "/Users/shiqi/projects/sandbox"
	out, err := cmd.Output()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LString(out))
	return 1
}

func assertAccountStates(L *lua.LState) int {
	account := L.CheckString(1)
	expected := L.CheckTable(2)
	algodClient := getAlgodClient()
	resp, err := algodClient.AccountInformation(account).Do(context.Background())
	if err != nil {
		fmt.Printf("error getting account info: %v\n", err)
		return 0
	}

	// check balance
	if expected.RawGet(lua.LString("balance")) != nil {
		bal, _ := strconv.ParseInt(expected.RawGet(lua.LString("balance")).String(), 10, 64)
		if uint64(bal) != resp.Amount {
			fmt.Printf("Error: account balance %d does not match expected balance %d", resp.Amount, bal)
			return 0
		}
	}

	//check apps count
	if expected.RawGet(lua.LString("totalapps")) != nil {
		appCounts, _ := strconv.ParseInt(expected.RawGet(lua.LString("totalapps")).String(), 10, 64)
		if uint64(appCounts) != resp.TotalCreatedApps {
			fmt.Printf("Error: total created apps %d is not %d", len(resp.CreatedApps), appCounts)
			return 0
		}
	}

	// check app params
	if expected.RawGet(lua.LString("apps")) != nil {
		apps := expected.RawGet(lua.LString("apps"))
		//apps.
		appsJSON, _ := luajson.Encode(apps)
		var appsArr []map[string]interface{}
		json.Unmarshal(appsJSON, &appsArr)
		//fmt.Println(apps)
		//fmt.Println(string(val))
		//fmt.Printf("%+v\n", table)

		actualApps := make(map[uint64]models.Application)
		for _, app := range resp.CreatedApps {
			actualApps[app.Id] = app
		}
		for i := 0; i < len(appsArr); i++ {
			expectedApp := appsArr[i]
			actualApp := actualApps[uint64(expectedApp["id"].(float64))]
			extraPages := uint64(expectedApp["extrapages"].(float64))
			if extraPages != actualApp.Params.ExtraProgramPages {
				fmt.Printf("Error: app extra pages %d is not %d\n", actualApp.Params.ExtraProgramPages, extraPages)
				return 0
			}
		}
	}
	return 0
}

func getKMDClient() kmd.Client {
	kmdClient, err := kmd.MakeClient("http://localhost:4002", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		fmt.Printf("failed to make kmd client: %s\n", err)
		return kmd.Client{}
	}
	return kmdClient
}
func getAlgodClient() *gosdk.Client {
	algodClient, err := gosdk.MakeClient("http://localhost:4001", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		fmt.Printf("failed to make algod client: %s\n", err)
		return nil
	}
	return algodClient
}
