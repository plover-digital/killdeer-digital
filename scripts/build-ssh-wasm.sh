#!/bin/sh
# Builds the browser terminal's SSH core into static/ssh/ for local
# development. The Containerfile does the same for deploys.
set -eu
cd "$(dirname "$0")/.."
(cd webssh && GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../static/ssh/killdeer-ssh.wasm .)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" static/ssh/wasm_exec.js
ls -l static/ssh/killdeer-ssh.wasm
