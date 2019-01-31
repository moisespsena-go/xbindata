#!/bin/bash

executable=bin/embed

echo "prepare assets..."
xbindata -embed -pkg assets -prefix assets_root -o assets/assets.go assets_root/... || exit $?

echo "build project..."
go build -o $executable || exit $?

echo "embed asset contents into executable"
go run assets/data_store/main.go $executable || exit $?

echo done.