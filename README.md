[![Travis](https://travis-ci.com/hdac-io/friday.svg?token=bhU3g7FdixBp5h3M2its&branch=dev)](https://travis-ci.com/hdac-io/friday/branches)
[![codecov](https://codecov.io/gh/hdac-io/friday/branch/dev/graph/badge.svg?token=hQEgzmULjh)](https://codecov.io/gh/hdac-io/friday)

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

# create a wallet key
clif keys add elsa # select password
clif keys add anna # select password

# add genesis node
nodef add-genesis-account $(clif keys show elsa -a) 1000dummy,100000000stake
nodef add-genesis-account $(clif keys show anna -a) 1000dummy,100000000stake
nodef add-el-genesis-account $(clif keys show elsa -a) "5000000000000" "1000000"
nodef add-el-genesis-account $(clif keys show anna -a) "5000000000000" "1000000"

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
* edit ~/.nodef/config/config.toml
```
...
# Maximum size of request body, in bytes
max_body_bytes = 1000000 -> 3000000
...
# Maximum size of a single transaction.
# NOTE: the max size of a tx transmitted over the network is {max_tx_bytes} + {amino overhead}.
max_tx_bytes = 1048576 -> 3145728
...
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
# Maximum size of request body, in bytes
max_body_bytes = 1000000 -> 3000000
...
# Comma separated list of seed nodes to connect to
seeds = "" -> "<genesis node's ID>@<genesis node's IP>:26656"
...
# Maximum size of a single transaction.
# NOTE: the max size of a tx transmitted over the network is {max_tx_bytes} + {amino overhead}.
max_tx_bytes = 1048576 -> 3145728
...
```
* replace `~/.nodef/config/genesis.json` to genesis node's one what you saved above.
## Test

```
$ make test
```
