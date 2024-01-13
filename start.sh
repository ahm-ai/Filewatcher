#!/bin/bash

# Set environment variables
export WATCH_PATH="~/Documents/Extension/apps/server"
export WATCH_REGEX=".*\\.js$"
export COMMAND_NAME="node ws_server.js"

# Run your Go program with these environment variables
go run main.go --path="$WATCH_PATH" --regex="$WATCH_REGEX" --command="$COMMAND_NAME"
