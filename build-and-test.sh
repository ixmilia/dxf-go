#!/bin/sh -e

go version
go generate
go build -v
go test -v
