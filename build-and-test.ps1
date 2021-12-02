#!/usr/bin/pwsh

Set-StrictMode -version 2.0
$ErrorActionPreference = "Stop"

function Fail([string]$message) {
    throw $message
}

try {
    go version || Fail "Error reporting `go` version"
    go generate || Fail "Error generating code"
    go build -v || Fail "Error building library"
    go test -v || Fail "Error testing library"
    go build -v ./examples || Fail "Error building examples"
}
catch {
    Write-Host $_
    Write-Host $_.Exception
    Write-Host $_.ScriptStackTrace
    exit 1
}
