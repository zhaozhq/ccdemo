#!/bin/bash
set -e

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.."

# Auto-format all Go files
gofmt -w .

# Auto-fix imports if goimports is available
if command -v goimports > /dev/null 2>&1; then
    goimports -w .
fi
