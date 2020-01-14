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
nodef add-el-genesis-account $(clif keys show elsa -p) "5000000000000" "1000000"
nodef add-el-genesis-account $(clif keys show anna -p) "5000000000000" "1000000"
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
#### Readable ID service

Hdac main supports readable ID for better usage. You may organize up to **20 letters** with 0-9, a-z, '-', '.', and '_' .
With this feature, you don't have to memo recipient's complex hashed address. Just remember easy address and send token!
Of course, you can also use previous hashed address system. This is optional for your availability.

* Set readable ID to account
  You may register readable ID in two types: **Bech32 public key** or **Secp256k1 public key**.
  (These **Bech32 public key** and **Secp256k1 public key** are convertible in both-way. One's encoding is the other's decoding.)
  * Usage:
    * By Bech32 public key: `clif readableid setbybech32pubkey princesselsa $(clif keys show elsa -p)`
    * By Secp256k1 public key: `clif readableid setbypubkey princesselsa 03c676be88d995c130f9a1bbe43fac18ab1db3b1173e6681bf4741bb8fb9f2bbbf`
```bash
clif readableid setbybech32pubkey princesselsa $(./clif keys show elsa -p)

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
    (CAREFUL: `elsa` is an alias of your local wallet, **not readable ID**. The name of the local wallet alias is not stored in mainnet.)

* Check readable ID mappint status
  * Usage: `clif readableid getaccount princesselsa`
```bash
clif readableid getaccount princesselsa

{
  "name": "princesselsa",
  "address": "friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv",
  "pubkey": "A7TB68o82IU3jpP0oiiGVtN9glfWtyWiBE+k4gFWMq3Q"
}
```

* Change public key of readable ID
  * Usage
    * By Bech32 public key: `clif readableid changekeybech32 princesselsa $(./clif keys show elsa -p) $(./clif keys show anotherelsa -p)`
    * By Secp256k1 public key: `clif readableid changekey princesselsa 03c676be88d995c130f9a1bbe43fac18ab1db3b1173e6681bf4741bb8fb9f2bbbf 02553a5b6ac99ed7a8b5dbde40738b61b76a8192d0f3b71cf8c65a34ff7909bd2f`

```bash
clif readableid changekeybech32 princesselsa $(./clif keys show elsa -p) $(./clif keys show anotherelsa -p)

Change key for readable name  bryan
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
        "old_pubkey": "A3CcPTCO6Jr9Q0HSlgOK4unk+atk4n71AK2J+sQqkKy1",
        "new_pubkey": "A7TB68o82IU3jpP0oiiGVtN9glfWtyWiBE+k4gFWMq3Q"
      }
    }
  ],
  "memo": ""
}
```

#### Operation with readable ID
* Query
  * Usage: `clif executionlayer getbalance [--name or --pubkey or --fridaypub]`
```bash
clif executionlayer getbalance --name princesselsa
clif executionlayer getbalance --fridaypub $(clif keys show elsa -a)

{
   "value": "5000000000000"
}
```

* transfer (send)
  * usage: `clif executionlayer transfer [--token_contract_address] [--from local_wallet_alias] [--to-name readable_id or --to-pubkey secp256k1 pubkey or --to-fridaypub bech32 pubkey] [--amount] [--fee] [--gas-price]`
  * Flag `--token_contract_address` is currently dummy. You may input as same as `from_address`
```bash
clif executionlayer transfer --from elsa --to-name sisteranna 1000000 100000000 20000000

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
  * usage: `clif executionlayer bond [--from local_wallet_alias] [--validator bech32 validator address "fridayvaloperxxxxxx..."] [--amount] [--fee] [--gas-price]`
```sh
./clif executionlayer bond \
--from elsa \
--validator $(clif keys show elsa --bech val -a) \
--amount 1000000 \
--fee 100000000 \
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
  * usage: `clif executionlayer unbond [--from local_wallet_alias] [--validator bech32 validator address "fridayvaloperxxxxxx..."] [--amount] [--fee] [--gas-price]`
```sh
./clif executionlayer unbond \
--from elsa \
--validator $(clif keys show elsa --bech val -a) \
--amount 1000000 \
--fee 100000000 \
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
* copy `~/.nodef/config/manifest.toml` to manifest node's one what you saved above.

### Running validator
* run on this fullly synchronized node
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
--fee 100000000 \
--gas-price 30000000 \
```

## Test

```
# run execution engine grpc server
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock

# run test
make test
```
