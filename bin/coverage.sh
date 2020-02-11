#! /bin/bash
set -e

go test -v -coverpkg=./internal/... -covermode=count -coverprofile=coverage.out ./test/...

go tool cover -html=coverage.out -o coverage.html