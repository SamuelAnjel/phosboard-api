#!/bin/bash
# Lint script for phosboard-backend

set -e

echo "Running golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.5.0
fi

golangci-lint run ./...

echo "✅ Lint passed!"