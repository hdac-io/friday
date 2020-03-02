#!/usr/bin/env bash

set -e

if [ ${PWD##*/} != "friday" ]; then
  echo "error : run this script in repository root"
  exit 1
fi

TARGET_BRANCH="master"
COMMIT_HASH="2184a48c28c1c048131d3a8f2f4b127ad9bad2c8"
if [ ! -d "CasperLabs/.git" ]; then
  git clone --single-branch --branch $TARGET_BRANCH https://github.com/hdac-io/CasperLabs.git
fi

cd CasperLabs
git fetch origin $TARGET_BRANCH
git reset --hard $COMMIT_HASH

cd execution-engine
make setup
cargo build --release # build execution engine

declare -a TARGET_CONTRACTS=(
  "mint-install"
  "pos-install"
  "counter-call"
  "counter-define"
  "bonding"
)

declare -a WASM_FILES=(
  "mint_install.wasm"
  "pos_install.wasm"
  "counter_call.wasm"
  "counter_define.wasm"
  "bonding.wasm"
)

for pkg in "${TARGET_CONTRACTS[@]}"; do
  make build-contract-rs/$pkg
done

CONTRACT_DIR="$HOME/.nodef/contracts"
mkdir -p $CONTRACT_DIR

for wasm in "${WASM_FILES[@]}"; do
  cp "./target/wasm32-unknown-unknown/release/$wasm" "$CONTRACT_DIR"
done
