# FRIDAY

[![Travis](https://travis-ci.com/hdac-io/friday.svg?token=bhU3g7FdixBp5h3M2its&branch=master)](https://travis-ci.com/hdac-io/friday/branches)
[![codecov](https://codecov.io/gh/hdac-io/friday/branch/master/graph/badge.svg?token=hQEgzmULjh)](https://codecov.io/gh/hdac-io/friday)
[![license](https://img.shields.io/github/license/hdac-io/friday.svg)](https://github.com/hdac-io/friday/blob/master/LICENSE)

Welcome to the official FRIDAY repository.
Friday is a decentralized network program which helps you to connect other blockchain and build your own network easily.

## Build the source

## Supported Systems

We currently supports the operating systems below.

* Ubuntu 18.04 or later
* MacOS 10.14 or later

## Prerequisites

You should install the packages below before you build the source.

* [Rust](https://www.rust-lang.org/tools/install)
* [Golang](https://golang.org/doc/install) >= 1.13
* [protoc](http://google.github.io/proto-lens/installing-protoc.html) >= 3.6.1
* make
* cmake

## Build

git clone this source and change directory.

```sh
git clone https://github.com/hdac-io/friday.git
cd friday
```

Simply make it!

```sh
make install
```

The built binaries - _nodef and clif_ - will be located in your `$GOBIN`.

## Test

You should launch execution engine grpc server first.

```sh
cd friday
./CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server -z $HOME/.casperlabs/.casper-node.sock&
```

And simply make it again!

```sh
make test
```

## Documents

* [Tutorials](https://docs.hdac.io/first-step/installation)
* [Validator guide](https://docs.hdac.io/validators/become-a-validator)
* [CLI usage](https://docs.hdac.io/cli/nickname)
* [Restful API usage](https://docs.hdac.io/restful-api/block-tx)
* [Release log](https://docs.hdac.io)

## Resources

* [Official Site](https://hdactech.com)
* [Forum](https://forum.hdac.io)
* Medium: [Eng](https://medium.com/hdac) / [Kor](https://medium.com/hdackorea)

## License

Friday is licensed under the [Apache License 2.0](https://github.com/hdac-io/friday/blob/master/LICENSE).
