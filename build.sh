#!/bin/sh

repo='https://github.com/pingcap/index_advisor'

case $(uname -s) in
    Linux|linux) os=linux ;;
    Darwin|darwin) os=darwin ;;
    *) os= ;;
esac

if [ -z "$os" ]; then
    echo "OS $(uname -s) not supported." >&2
    exit 1
fi

case $(uname -m) in
    amd64|x86_64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) arch= ;;
esac

if [ -z "$arch" ]; then
    echo "Architecture  $(uname -m) not supported." >&2
    exit 1
fi

GITHASH=$(git rev-parse HEAD)
BIN_OUTPUT="index-advisor-$os-$arch"

GO111MODULE=on go build -ldflags "-X github.com/pingcap/index_advisor/version.GitHash=$GITHASH" -o bin/$BIN_OUTPUT main.go

tar -czf bin/$BIN_OUTPUT.tar.gz bin/$BIN_OUTPUT