#!/usr/bin/env bash

set -e

if [ ${PWD##*/} != "friday" ]; then
  echo "error : run this script in repository root"
  exit 1
fi

COMMIT_HASH="0a928aad5ede6b7dbadc42dd858b6b740a4deba4"
if [ ! -d "CasperLabs/.git" ]; then
  git clone https://github.com/hdac-io/CasperLabs.git
fi

cd CasperLabs
git fetch origin
git reset --hard $COMMIT_HASH

cd execution-engine
make setup-rs
cargo build --release # build execution engine

declare -a TARGET_CONTRACTS=(
  "hdac-mint-install"
  "pop-install"
  "counter-call"
  "counter-define"
  "bonding"
  "standard-payment-install"
)

declare -a WASM_FILES=(
  "hdac_mint_install.wasm"
  "pop_install.wasm"
  "counter_call.wasm"
  "counter_define.wasm"
  "bonding.wasm"
  "standard_payment_install.wasm"
)

for pkg in "${TARGET_CONTRACTS[@]}"; do
  make build-contract-rs/$pkg
done

CONTRACT_DIR="$HOME/.nodef/contracts"
mkdir -p $CONTRACT_DIR

for wasm in "${WASM_FILES[@]}"; do
  cp "./target/wasm32-unknown-unknown/release/$wasm" "$CONTRACT_DIR"
done
