[![Travis](https://travis-ci.com/hdac-io/friday.svg?token=bhU3g7FdixBp5h3M2its&branch=master)](https://travis-ci.com/hdac-io/friday/branches)
[![codecov](https://codecov.io/gh/hdac-io/friday/branch/master/graph/badge.svg?token=hQEgzmULjh)](https://codecov.io/gh/hdac-io/friday)

# TESTNET

## Prerequisites

* [Rust](https://www.rust-lang.org/tools/install)
* [Golang](https://golang.org/doc/install) >= 1.13
* [protoc](http://google.github.io/proto-lens/installing-protoc.html) >= 3.6.1
* make and cmake

## Build

```
make install
```

## Run

### Setup a genesis status and run a genesis node
* note: Fill the name what you want inside < >
```sh
# run execution engine grpc server
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock

# init node
nodef init <node_name> --chain-id testnet

# copy execution engine chain configurations
cp ./x/executionlayer/resources/manifest.toml ~/.nodef/config

# create a wallet key
clif keys add elsa # select password
clif keys add anna # select password

# add genesis node
nodef add-genesis-account $(clif keys show elsa -a) 5000000000000dummy,100000000stake
nodef add-genesis-account $(clif keys show anna -a) 5000000000000dummy,100000000stake
nodef add-el-genesis-account $(clif keys show elsa -a) "5000000000000" "1000000"
nodef add-el-genesis-account $(clif keys show anna -a) "5000000000000" "1000000"
nodef load-chainspec ~/.nodef/config/manifest.toml

# apply default clif configure
clif config chain-id testnet
clif config output json
clif config indent true
clif config trust-node true

# prepare genesis status
nodef gentx --name elsa # insert password
nodef collect-gentxs
nodef validate-genesis
```
* genesis node start
```
nodef start
```
* note your genesis node's ID and genesis file
* you can get your node ID using clif
```
clif status | grep \"id\"
```
* your genesis file exists `~/.nodef/config/genesis.json`

### Clif usage
#### Nickname service for readability

Hdac mainnet supports readable ID for better usage. You may organize up to **20 letters** with 0-9, a-z, '-', '.', and '_' .
With this feature, you don't have to memo recipient's complex hashed address. Just remember easy address and send token!
Of course, you can also use previous hashed address system. This is optional for your availability.

* Set readable ID to account
  You may register by address, and the address can be access from two ways: wallet alias and address (e.g. friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv)
  * Usage:
    * By address: `clif nickname set princesselsa --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv`
    * By local wallet alias: `clif nickname set princesselsa --wallet walletelsa`
```bash
clif nickname set princesselsa --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv
clif nickname set princesselsa --wallet walletelsa

{
  "chain_id": "testnet",
  "account_number": "1",
  "sequence": "2",
  "fee": {
    "amount": [],
    "gas": "200000"
  },
  "msgs": [
    {
      "type": "readablename/SetName",
      "value": {
        "name": {
          "H": "1183736206936",
          "L": "0"
        },
        "address": "friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv",
        "pubkey": "AiJmIrS9ZPdCWmzQ92BZUxzJ49eGdF0aTCPw60a+Ft/2"
      }
    }
  ],
  "memo": ""
}
```
  * After confirmation, you may use `princesselsa` as a readable ID instead of `friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv`
    (CAREFUL: `walletelsa` is an alias of your local wallet, **not your nickname**. The name of the local wallet alias is not stored in mainnet.)

* Check readable ID mappint status
  * Usage: `clif nickname get-address princesselsa`
```bash
clif nickname get-address princesselsa

{
  "name": "princesselsa",
  "address": "friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv"
}
```

* Change address of nickname
  * Usage
    * By wallet alias: `clif nickname change-to princesselsa friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz --wallet walletanna`
    * By address directly: `clif nickname change-to princesselsa friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv`

```bash
clif nickname change-to princesselsa friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz --wallet walletanna
clif nickname change-to princesselsa friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv

friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv  ->  friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz
{
  "chain_id": "testnet",
  "account_number": "6",
  "sequence": "0",
  "fee": {
    "amount": [],
    "gas": "200000"
  },
  "msgs": [
    {
      "type": "readablename/ChangeKey",
      "value": {
        "ID": "bryan",
        "old_address": "friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv",
        "new_address": "friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz",
      }
    }
  ],
  "memo": ""
}
```

#### Operation with nickname
*TODO: Currently described only Hdac custom CLI. Need to add for general cases*

* Query
  * Usage: `clif hdac getbalance --wallet|--nickname|--address <owner>`
```bash
clif hdac getbalance --wallet walletelsa
clif hdac getbalance --nickname princesselsa
clif hdac getbalance --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv

{
   "value": "5000000000000"
}
```

* transfer (send)
  * usage: `clif hdac transfer-to <recipient_address_or_nickname> <amount> <fee> <gas-price> --address|--wallet|--nickname <from>`
```bash
clif hdac transfer-to sisteranna 1000000 100000000 20000000 --wallet walletelsa
clif hdac transfer-to sisteranna 1000000 100000000 20000000 --nickname princesselsa
clif hdac transfer-to sisteranna 1000000 100000000 20000000 --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv

...
confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'walletelsa': # input your password
{
  "height": "0",
  "txhash": "141F12A891659F52B055EF7F701B1D406E5F1721CE929630CC5CE3CE0C4C8718",
  "raw_log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"executionengine\"}]}]}]",
  "logs": [
    {
      "msg_index": 0,
      "success": true,
      "log": "",
      "events": [
        {
          "type": "message",
          "attributes": [
            {
              "key": "action",
              "value": "executionengine"
            }
          ]
        }
      ]
    }
  ]
}
```
* bond
  * usage: `clif hdac bond --wallet|--address|--nickname <owner> <amount> <fee> <gas-price>`
```sh
clif hdac bond --wallet walletelsa 1000000 100000000 30000000
clif hdac bond --nickname princesselsa 1000000 100000000 30000000
clif hdac bond --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv 1000000 100000000 30000000

confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'walletelsa':
{
  "height": "0",
  "txhash": "22DF1E0D8D9EB8BE2B5F50995C6FC0AB20E34715875A9F9856A9466A8C406807",
  "raw_log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"executionengine\"}]}]}]",
  "logs": [
    {
      "msg_index": 0,
      "success": true,
      "log": "",
      "events": [
        {
          "type": "message",
          "attributes": [
            {
              "key": "action",
              "value": "executionengine"
            }
          ]
        }
      ]
    }
  ]
}
```

* unbond
  * usage: `clif hdac unbond --wallet|--address|--nickname <owner> <amount> <fee> <gas-price>`
```sh
clif hdac unbond --wallet walletelsa 1000000 100000000 30000000
clif hdac unbond --nickname princesselsa 1000000 100000000 30000000
clif hdac unbond --address friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv 1000000 100000000 30000000

confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'walletelsa':
{
  "height": "0",
  "txhash": "69C51D25E3E5DB4F2D4ACE832C775DC8EE993E9CDB7560A3AF470FF07CC7FFC9",
  "raw_log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"executionengine\"}]}]}]",
  "logs": [
    {
      "msg_index": 0,
      "success": true,
      "log": "",
      "events": [
        {
          "type": "message",
          "attributes": [
            {
              "key": "action",
              "value": "executionengine"
            }
          ]
        }
      ]
    }
  ]
}
```

### Connect to seed node
* run this on another machine
```sh
# run execution engine grpc server
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock

# init node
nodef init <node_name> --chain-id testnet
```
* edit ~/.nodef/config/config.toml
```
...
# Comma separated list of seed nodes to connect to
seeds = "" -> "<genesis node's ID>@<genesis node's IP>:26656"
...
```
* replace `~/.nodef/config/genesis.json` to genesis node's one what you saved above.
* copy `~/.nodef/config/manifest.toml` to manifest node's one what you saved above.

### Running validator
* run on this fullly synchronized node
* create a wallet key
```sh
clif keys add bryan # select password
```

* create validator
```sh
nodef tendermint show-validator
# fridayvalconspub16jrl8jvqq929y3r2dem455nptpd9g3mn0929q5eswaay6365vdtrx6j42dkrxtek24n5ycmpfax9s4mp9apkgkpe2vux64e0xe3xz5f09ucrje6e25cxwe3tf3kxjc6gfesnyv308p382ujc24snqn2kwfq45j60gc6nqs6wvfp8xen3d3ersnjnxfmrv6jv8pjxsmjtv3kxcapc09y5w5sa9v92q

 clif hdac create-validator \
--from bryan \
--pubkey fridayvalconspub16jrl8jvqq929y3r2dem455nptpd9g3mn0929q5eswaay6365vdtrx6j42dkrxtek24n5ycmpfax9s4mp9apkgkpe2vux64e0xe3xz5f09ucrje6e25cxwe3tf3kxjc6gfesnyv308p382ujc24snqn2kwfq45j60gc6nqs6wvfp8xen3d3ersnjnxfmrv6jv8pjxsmjtv3kxcapc09y5w5sa9v92q \
--moniker valiator-bryan

# or --pubkey $(nodef tendermint show-validator)
```
* bonding amount
```sh
clif hdac bond --wallet walletelsa 1000000 100000000 30000000
```

## Test

```
# run execution engine grpc server
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock

# run test
make test
```
