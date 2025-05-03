# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands
- Build binary: `go build -o github-mcp-server ./cmd/github-mcp-server`
- Run unit tests: `go test ./pkg/...`
- Run specific test: `go test -run TestName ./pkg/path`
- Run e2e tests: `GITHUB_MCP_SERVER_E2E_TOKEN=<TOKEN> go test -v --tags e2e ./e2e`
- Format code: `go fmt ./...`

## Code Style Guidelines
- Use Go 1.23+ standard formatting conventions
- Error handling: Always check errors and return them with context
- Helper functions for parameter validation in pkg/github/server.go
- Testing: Use testify for assertions and mocks (github-mock for HTTP)
- Always include unit tests for new functionality
- Function documentation follows Go standard format with descriptions
- Use typed parameters with generics when appropriate
- Pagination: Default page=1, perPage=30 (max 100)
- Imports should be grouped: standard lib, then external packages