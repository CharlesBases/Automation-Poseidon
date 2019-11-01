#!/bin/zsh

set -e

export GO111MODULE=off

go get github.com/golang/protobuf/{proto,protoc-gen-go}
go get github.com/gogo/protobuf/proto
go get github.com/micro/micro
go get github.com/micro/protoc-gen-micro
go install github.com/gogo/protobuf/protoc-gen-gogofaster

# mac
# brew install protobuf

# linux


EXPORT GO111MODULE=on