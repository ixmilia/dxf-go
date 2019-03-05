@echo off

go build -v
if errorlevel 1 goto :error

goto :eof

:error
echo Error building examples
exit /b 1
