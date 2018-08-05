#!/bin/bash

set -ex

./build.sh arm
cp ./strut ../../../../../pi-gen/stage3/00-install-packages/files/
