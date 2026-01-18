#!/bin/bash

# build and run taskbox

echo "building taskbox..."
go build -o taskbox ./cmd/server

if [ $? -eq 0 ]; then
	echo "build successful"
	echo "starting server on http://localhost:1234"
	./taskbox
else
	echo "build failed"
	exit 1
fi
