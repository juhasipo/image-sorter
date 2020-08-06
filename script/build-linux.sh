#!/bin/bash

mkdir -p out/linux
go build -v -i -o out/linux ./...
