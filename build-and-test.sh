#!/bin/sh -e

# logging
go version

# main task
go generate
go build -v
go test -v

# verify examples
cd examples
./build.sh
