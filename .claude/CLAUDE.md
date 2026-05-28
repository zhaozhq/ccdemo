# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A minimal Go module (`hello`) using Go 1.26.1. Entry point is [main.go](main.go).

## Common Commands

- **Run**: `go run main.go`
- **Build**: `go build`
- **Tidy dependencies**: `go mod tidy`

## Project Permissions

The local `.claude/settings.json` auto-allows `go mod *` and `go run *` commands.
