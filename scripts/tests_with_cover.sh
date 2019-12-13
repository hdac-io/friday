#!/usr/bin/env bash
set -e
echo "" > coverage.txt
for d in $(go list ./...); do
  go test -v -race -coverprofile=profile.out $d
  if [ -f profile.out ]; then
    cat profile.out | grep -v "client/cli" | grep -v "client/rest" | grep -v "friday/cmd" >> coverage.txt
    rm profile.out
  fi
done