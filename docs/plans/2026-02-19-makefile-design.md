# Makefile Design

**Date:** 2026-02-19
**Status:** Approved

## Summary

Add a single flat `Makefile` to the project root with common development directives for building, testing, linting, formatting, cleaning, releasing, and installing the `imgutil` binary.

## Variables

| Variable | Value |
|---|---|
| `BINARY` | `imgutil` |
| `MODULE` | `github.com/thisisnotashwin/imgutil` |
| `VERSION` | `git describe --tags --always --dirty`, fallback to `dev` |
| `LDFLAGS` | `-X $(MODULE)/internal/version.Version=$(VERSION)` |
| `PLATFORMS` | `linux/amd64 linux/arm64 darwin/amd64 darwin/arm64` |

## Targets

| Target | Description |
|---|---|
| `help` | Default target — prints annotated targets from `## comments` |
| `build` | Compile binary for current platform → `./bin/imgutil` |
| `test` | Run all tests with race detector (`go test -race ./...`) |
| `lint` | Run `golangci-lint run` |
| `fmt` | Run `gofmt -w .` and `go vet ./...` |
| `clean` | Remove `./bin/` and `./dist/` |
| `release` | Cross-compile for all platforms → `./dist/imgutil-<os>-<arch>` |
| `install` | Install binary to `$GOPATH/bin` via `go install` |

## Conventions

- `.PHONY` declared for all non-file targets
- `@` prefix on commands to suppress echo noise
- `help` is the default (first) target — bare `make` prints usage
- Self-documenting: targets annotated with `## description` comments
