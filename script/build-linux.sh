#!/bin/bash

# Change version for this build
VERSION="$(cat VERSION)-$(git rev-parse --short HEAD)"
sed -i "s/__IMAGE_SORTER_VERSION__/${VERSION}/g" common/constants.go

mkdir -p out/linux
go build -ldflags "-s -w" -v -i -o out/linux ./...

# Revert the version back to normal
sed -i "s/${VERSION}/__IMAGE_SORTER_VERSION__/g" common/constants.go
