@echo off

go version
if errorlevel 1 goto :error

go generate
if errorlevel 1 goto :error

go build -v
if errorlevel 1 goto :error

go test -v
if errorlevel 1 goto :error
goto :eof

:error
echo Error building/testing
exit /b 1
