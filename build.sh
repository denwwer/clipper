#!/bin/bash

echo "Building..."
rm -rf build
rm -rf bin
mkdir build
mkdir bin

GOOS=linux GOARCH=amd64 go build -o bin/cliper main.go

echo "Zipping..."
ZIP_NAME="cliper.zip"
NAME="build/$ZIP_NAME"
zip $NAME bin/cliper config.gcfg -j
