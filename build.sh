#!/bin/bash

set -ex

case "$1" in
  arm)
    export GOOS=linux
    export GOARCH=arm
    export GOARM=7
  ;;
  vagrant)
    export GOOS=linux
    export GOARCH=amd64
  ;;
esac

rm -f strut
exec go build -ldflags "-X main.BuildVersion=$(date +%s) -X main.Version=0.0.1" -o strut -v .
