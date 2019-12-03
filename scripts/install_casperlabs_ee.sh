#!/usr/bin/env bash

set -e

if [ ${PWD##*/} != "friday" ]; then
  echo "error : run this script in repository root"
  exit 1
fi

NODE_DIR="$HOME/.nodef"

rm -rf CasperLabs $NODE_DIR

git clone --single-branch --branch release-v0.9 https://github.com/CasperLabs/CasperLabs

cd CasperLabs/execution-engine
make setup
cargo build --release # build execution engine

declare -a TARGET_CONTRACTS=(
  "mint-install"
  "pos-install"
  "counter-call"
  "counter-define"
  "standard-payment"
  "transfer-to-account"
  "bonding"
  "unbonding"
)

declare -a WASM_FILES=(
  "mint_install.wasm"
  "pos_install.wasm"
  "counter_call.wasm"
  "counter_define.wasm"
  "standard_payment.wasm"
  "transfer_to_account.wasm"
  "bonding.wasm"
  "unbonding.wasm"
)

for pkg in "${TARGET_CONTRACTS[@]}"; do
  make build-contract/$pkg
done

CONTRACT_DIR="$NODE_DIR/contracts"
mkdir -p $CONTRACT_DIR

for wasm in "${WASM_FILES[@]}"; do
  cp "./target/wasm32-unknown-unknown/release/$wasm" "$CONTRACT_DIR"
done
