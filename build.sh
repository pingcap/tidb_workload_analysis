#!/bin/sh

function compile()
{
  echo "build for $1-$2"
  GITHASH=$(git rev-parse HEAD)
  GO111MODULE=on CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -ldflags "-X github.com/pingcap/index_advisor/version.GitHash=$GITHASH" -o bin/index-advisor main.go
  cd bin
  tar -czf index-advisor-$1-$2.tar.gz index-advisor
  cd ..
}

OS=darwin
ARCH=amd64
compile $OS $ARCH

OS=darwin
ARCH=arm64
compile $OS $ARCH

OS=linux
ARCH=amd64
compile $OS $ARCH

OS=linux
ARCH=arm64
compile $OS $ARCH

