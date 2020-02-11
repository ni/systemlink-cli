#! /bin/bash
set -e

packages=$(go list ./internal/... | grep -v "internal/ssh" | tr '\n' ',')
go test -v -coverpkg=$packages -covermode=count -coverprofile=coverage.out ./test/...

go tool cover -html=coverage.out -o coverage.html