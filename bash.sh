#!/bin/zsh

set -e

name=proto

go build main.go

gopath=$GOPATH
gopath=${gopath%:*}

mv main ${gopath}/bin/${name}