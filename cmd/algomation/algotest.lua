local algotest = require("algotest")
--create first account
acct1=algotest.makeAccount()
print("acct1: ", acct1)
--create an app using configs in demo.yml
txn, contractAddr = algotest.createAppFromConfig(acct1,"contract1")
print("contract addr: ",contractAddr)
print("starting private network")
algotest.startPrivateNetwork()
-- sumbit app create txn
appID=txn:submit("533e921d0920c6f82d2874daefa70f76") -- walletID.
print("created app id:", appID)
-- assert states
expected={balance=999000,apps={{id=appID,extrapages=1}},totalapps=1}
algotest.assertAccountStates(acct1,expected)
-- create another account
acct2=algotest.makeAccount()
-- assert
expected={balance=1000000}
algotest.assertAccountStates(acct2,expected)
