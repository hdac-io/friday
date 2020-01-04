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
* query
  * usage: `clif executionlayer getbalance [address]`
```
clif executionlayer getbalance $(clif keys show elsa -a)

{
   "value": "5000000000000"
}
```

* transfer (send)
  * usage: `clif executionlayer transfer [token_contract_address] [from_address] [to_address]  [amount] [fee] [gas_price]`
  * `token_contract_address` is currently dummy, and you may input as same as `from_address`
```sh
clif executionlayer transfer $(clif keys show elsa -a) $(clif keys show elsa -a) $(clif keys show anna -a) 1000000 100000000 20000000

...
confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'elsa': # input your password
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
  * usage: `clif executionlayer bond`
```sh
./clif executionlayer bond \
--from $(clif keys show elsa -a) \
--validator fridayvaloper19rxdgfn3grqgwc6zhyeljmyas3tsawn64dsges \
--amount 1000000 \
--fee 10000000 \
--gas-price 30000000

confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'elsa':
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
  * usage: `clif executionlayer unbond [from_address] [unbond_amount] [fee] [gas_price]`
```sh
./clif executionlayer unbond \
--from $(clif keys show elsa -a) \
--validator fridayvaloper19rxdgfn3grqgwc6zhyeljmyas3tsawn64dsges \
--amount 1000000 \
--fee 10000000 \
--gas-price 30000000

confirm transaction before signing and broadcasting [y/N]: y
Password to sign with 'elsa':
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

### Running validator
* run this on another machine
* create a wallet key
```sh
clif keys add bryan # select password
```
* show AccAddress & ValAddress
```sh
# AccAddress
clif keys show bryan --bech acc

{
  "name": "bryan",
  "type": "local",
  "address": "friday19rxdgfn3grqgwc6zhyeljmyas3tsawn6qe0quc",
  "pubkey": "fridaypub1addwnpepqfaxrvy4f95duln3t6vvtd0qd0sdpwfsn3fh9snpnq06w25qualj6rxm04t"
}

# ValAddress
clif keys show bryan --bech val

{
  "name": "bryan",
  "type": "local",
  "address": "fridayvaloper19rxdgfn3grqgwc6zhyeljmyas3tsawn64dsges",
  "pubkey": "fridayvaloperpub1addwnpepqfaxrvy4f95duln3t6vvtd0qd0sdpwfsn3fh9snpnq06w25qualj6vczad0"
}
```
* create validator
```sh
 clif executionlayer create-validator \
--from=friday19rxdgfn3grqgwc6zhyeljmyas3tsawn6qe0quc \
--pubkey=$(nodef tendermint show-validator) \
--moniker=bryan
```
* bonding amount
```sh
clif executionlayer bond \
--from friday19rxdgfn3grqgwc6zhyeljmyas3tsawn6qe0quc \
--validator fridayvaloper19rxdgfn3grqgwc6zhyeljmyas3tsawn64dsges \
--amount 1000000 \
--fee 10000000 \
--gas-price 30000000 \
```

## Test

```
# run execution engine grpc server
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock

# run test
make test
```
