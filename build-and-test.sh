#!/bin/sh -e

# logging
go version

# main task
go generate
go build -v
go test -v

# verify examples
go build -v ./examples
