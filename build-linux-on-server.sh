#!/usr/bin/env bash
set -euo pipefail

export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOSUMDB="${GOSUMDB:-off}"
export CGO_ENABLED=1

GOPROXY=off GOSUMDB=off go build -mod=vendor -o gophish .

echo "Build complete: ./gophish"
file ./gophish || true
