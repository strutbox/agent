#!/bin/bash

set -ex

case "$1" in
  arm)
    export GOOS=linux
    export GOARCH=arm
    export GOARM=7
    bootstrap=https://strut.zone
    workdir=/var/lib/strut
  ;;
  vagrant)
    export GOOS=linux
    export GOARCH=amd64
  ;;
  *)
    bootstrap=http://127.0.0.1:8000
    workdir=.
  ;;
esac

rm -f strut
exec go build -ldflags "-X main.BuildVersion=$(date +%s) -X main.Version=0.0.1 -X main.BootstrapHost=${bootstrap} -X main.WorkingDir=${workdir}" -o strut -v .
