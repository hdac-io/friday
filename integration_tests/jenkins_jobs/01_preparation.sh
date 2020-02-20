#!/bin/bash

rm -rf $HOME/.nodef; rm -rf $HOME/.clif

cd $HOME/friday && git fetch && git checkout $TEST_BRANCH && \
git reset --hard origin/$TEST_BRANCH && git rebase

PATH="$HOME/.cargo/bin:$PATH" && make install
PATH="$PATH:$HOME/go/bin" && nodef init node1 --chain-id ci_testnet
