#!/bin/bash

daemonize -E BUILD_ID=dontKillMe -e $HOME/ee.err -o $HOME/ee.log -p $HOME/ee.pid $HOME/friday/CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock
daemonize -E BUILD_ID=dontKillMe -e $HOME/node.err -o $HOME/node.log -p $HOME/node.pid $HOME/go/bin/nodef start
