#!/bin/bash
if [ $(uname -s) = "Linux" ]
then
    echo "Building for (linux, amd64)..."
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-w -s" -race -o build/service logger.go
else
    echo "Building for (darwin, amd64)..."
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-w -s" -race -o build/service logger.go
fi
