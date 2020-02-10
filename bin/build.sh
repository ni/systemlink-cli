#! /bin/bash
set -e

# download tag and message service swagger.json
echo "Cloning systemlink-OpenAPI-documents repo"
mkdir -p build/models
git clone https://github.com/ni/systemlink-OpenAPI-documents.git 2> /dev/null || (cd systemlink-OpenAPI-documents ; git pull)

echo "Copying model definitions"
cp systemlink-OpenAPI-documents/tag/nitag.yml build/models/tags.yml
cp systemlink-OpenAPI-documents/message/nimessage.yml build/models/messages.yml

# download dependencies
echo "Downloading golang dependencies"
go get ./...

# build linux executable
echo "Building Linux x86 executable"
GOOS=linux GOARCH=386 go build -o build/systemlink cmd/main.go

# build windows executable
echo "Building Windows x86 executable"
GOOS=windows GOARCH=386 go build -o build/systemlink.exe cmd/main.go

# build mac-os executable
echo "Building MacOS executable"
GOOS=darwin GOARCH=amd64 go build -o build/systemlink.osx cmd/main.go

echo "Done!"