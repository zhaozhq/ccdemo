#!/bin/bash

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.."

echo "[hooks] ========== Go Quality Gate =========="

# Compile check
if go build ./... > /dev/null 2>&1; then
    echo "[hooks] go build ./...   PASS"
else
    echo "[hooks] go build ./...   FAIL — fix before next session"
fi

# Static analysis
if go vet ./... > /dev/null 2>&1; then
    echo "[hooks] go vet ./...     PASS"
else
    echo "[hooks] go vet ./...     FAIL — fix before next session"
fi

# Reminder for uncommitted Go files
modified=$(git diff --name-only | grep '\.go$' || true)
if [ -n "$modified" ]; then
    echo "[hooks] Modified Go files:"
    echo "$modified" | sed 's/^/           /'
    echo "[hooks] Reminder: run the following before commit"
    echo "[hooks]   go test ./..."
    echo "[hooks]   go test -race ./..."
    echo "[hooks]   go test -cover ./..."
else
    echo "[hooks] No uncommitted Go files"
fi
