[![Travis](https://travis-ci.com/hdac-io/friday.svg?token=bhU3g7FdixBp5h3M2its&branch=dev)](https://travis-ci.com/hdac-io/friday/branches)
[![codecov](https://codecov.io/gh/hdac-io/friday/branch/dev/graph/badge.svg?token=hQEgzmULjh)](https://codecov.io/gh/hdac-io/friday)

# TESTNET

## Prerequisites

* [Rust](https://www.rust-lang.org/tools/install)
* [Golang](https://golang.org/doc/install) >= 1.13
* [protoc](http://google.github.io/proto-lens/installing-protoc.html) >= 3.6.1

## Build

`$ make install`

## Run

```sh
$ nodef init <moniker> --chain-id namechain
$ clif keys add jack
$ clif keys add alice

$ nodef add-genesis-account $(clif keys show jack -a) 1000nametoken,100000000stake
$ nodef add-genesis-account $(clif keys show alice -a) 1000nametoken,100000000stake
$ nodef add-el-genesis-account $(clif keys show jack -a) "500000000" "1000000"
$ nodef add-el-genesis-account $(clif keys show alice -a) "500000000" "1000000"

$ clif config chain-id namechain
$ clif config output json
$ clif config indent true
$ clif config trust-node true

$ nodef gentx --name jack
$ nodef collect-gentxs
$ nodef validate-genesis
$ nodef start
```

## Test

`$ make test`
