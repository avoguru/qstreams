#!/bin/bash

# Build and run the server
echo "Building the server..."
go build -o bin/server ./server

echo "Starting the server..."
./bin/server
