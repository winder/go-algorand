# Goran

The **Al-goran-d** developer automation command provides utilities and integration points for building consistent access to features across different technology stacks.

## JSON-RPC Service
The vision for **goran** is grand, but we've started small. An RPC server with a targetted set of utilities is available for adding Algorand private networks to your automated test. Think of it as sandbox for CI.

### Usage
Start goran on port `1234`
```bash
~$ goran rpc-server --tmpDir /tmp/goran-tmp-dir/ --port 1234
```

Send commands to goran, here we start a private network:
```bash
~$ curl -s -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0",
  "method":"Network.Start",
  "params":{"Name": "example-network", "DevMode": true},
  "id":1}' http://0.0.0.0:1234/rpc | jq
{
  "jsonrpc": "2.0",
  "result": {
    "Token": "9021da30ff204d885c9c0dd61097ba1e8b13134b2390a995badf8bdd1af1f282",
    "AdminToken": "7769a6e93a0083c50fd901788b415fba25d1e05998cadddec393940700ad603f",
    "URL": "127.0.0.1:45385",
    "Accounts": [
      {
        "Account": "WCFMIX3Y5ECC2DCQMYCHN3Z32L42BDQRIBJTTSPRS2I4K4KZ32VLE76RU4",
        "Mnemonic": "lab person truly unit vacuum kidney crater grow climb attract identify tag concert lawn version elegant party elite soccer valve wine whisper usual about slow",
        "PrivateKey": "4NPo0ttO+OlRBmdYsQPh0l0X+FmeRwQFkpsRz33qB36wisRfeOkELQxQZgR27zvS+aCOEUBTOcnxlpHFcVneqg=="
      }
    ]
  },
  "id": 1
}
```

When finished, you can also destroy the network:
```bash
curl -s -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0",
  "method":"Network.Destroy",
  "params":{"Name": "example-network", "DevMode": true},
  "id":1}' http://0.0.0.0:1234/rpc|jq
{
  "jsonrpc": "2.0",
  "result": {
    "Name": "example-network"
  },
  "id": 1
}
```


### Commands

#### Network.Start
Start a new private network.

#### Network.Destroy
Destroy a private network and all related resources.

#### Network.Status
Fetch the network status by name.

#### List
List all of the active networks.

#### System.Health
Check that the RPC server is healthy.

#### System.Shutdown
Shutdown the RPC server remotely.
