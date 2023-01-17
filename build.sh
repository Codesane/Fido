#!/bin/bash

[ -d build ] && rm -r build
mkdir build

env GOOS=linux GOARCH=amd64 go build -o build/fido -v main.go process.go

