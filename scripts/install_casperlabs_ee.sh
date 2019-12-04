#!/usr/bin/env bash

set -e

if [ ${PWD##*/} != "friday" ]; then
  echo "error : run this script in repository root"
  exit 1
fi

CASPERLABS_TARGET_TAG="v0.9.0"
if [ ! -d "CasperLabs/.git" ]; then
  git clone --single-branch --branch $CASPERLABS_TARGET_TAG https://github.com/CasperLabs/CasperLabs.git
fi

cd CasperLabs
git fetch origin refs/tags/$CASPERLABS_TARGET_TAG:refs/tags/$CASPERLABS_TARGET_TAG
git checkout $CASPERLABS_TARGET_TAG

cd execution-engine
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

CONTRACT_DIR="$HOME/.nodef/contracts"
mkdir -p $CONTRACT_DIR

for wasm in "${WASM_FILES[@]}"; do
  cp "./target/wasm32-unknown-unknown/release/$wasm" "$CONTRACT_DIR"
done
