# imgutil

## Project
Go CLI tool for inspecting Docker/OCI images. Uses `google/go-containerregistry` (ggcr) and `cobra`.
Module: `github.com/thisisnotashwin/imgutil`

## Git Workflow
- Always create a new branch before making any changes — never commit directly to `main`

## Build & Test
- `make build` → `./bin/imgutil`
- `make test` → unit tests; `make lint` → golangci-lint; `make fmt` → gofmt
- `go build ./...` must always pass after edits

## Command Architecture
- All commands live in `commands/<name>.go`, registered in `commands/root.go`
- Output always via `cmd.OutOrStdout()`, never `os.Stdout`
- Source selection: `sourceFromFlags(flags)` — never hard-coded
- Format selection: `formatFromFlags(flags)`
- All rendering in `internal/format/output.go`, never in command files
- Errors: `fmt.Errorf("context: %w", err)` — no ignored errors

## Testing
- New commands need `commands/<name>_test.go`: success path + error path minimum
- Integration tests use `//go:build integration` build tag

## ggcr Patterns
- Prefer lazy accessors (`img.Digest()`, `img.Size()`) over full manifest fetches
- Only fetch layer content when explicitly needed — manifests/configs are cheap
