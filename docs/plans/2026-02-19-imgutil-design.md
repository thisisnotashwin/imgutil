# imgutil Design

**Date:** 2026-02-19
**Status:** Approved

## Overview

`imgutil` is a Go CLI tool for inspecting Docker images — both surface-level metadata and per-layer
breakdowns. It works with images from the local Docker daemon and from remote registries, without
requiring a full pull for remote inspection.

## Goals

- Inspect image config metadata (OS/arch, entrypoint, env, exposed ports, labels, created timestamp)
- Inspect per-layer breakdown (digest, size, creating command, files added/removed)
- Support both local Docker daemon and remote registries as image sources
- Human-readable output by default; JSON output via `--output json` for scripting
- Single binary with subcommands, structured so commands can be split into separate binaries later

## Non-Goals

- Vulnerability scanning or SBOM generation (out of scope for now)
- Container lifecycle management (run, stop, exec)
- Registry push/tag/publish operations

## Technology Choices

- **[google/go-containerregistry](https://github.com/google/go-containerregistry)** — single image
  abstraction for both local (daemon transport) and remote (registry transport) sources
- **[cobra](https://github.com/spf13/cobra)** — CLI framework for subcommands and flag inheritance
- **No Docker SDK dependency** — `ggcr`'s daemon transport handles local image access

## Project Structure

```
imgutil/
├── cmd/
│   └── imgutil/
│       └── main.go          # entry point
├── internal/
│   ├── image/
│   │   └── loader.go        # resolves images from daemon or registry
│   └── format/
│       └── output.go        # human-readable vs JSON rendering
├── commands/
│   ├── root.go              # cobra root command + global flags
│   ├── inspect.go           # `imgutil inspect <image>`
│   └── layers.go            # `imgutil layers <image>`
├── go.mod
└── go.sum
```

## Subcommands

### `imgutil inspect <image>`
Displays image config metadata: OS, architecture, entrypoint, CMD, env vars, exposed ports, labels,
working directory, created timestamp, total size.

### `imgutil layers <image>`
Displays a per-layer breakdown: layer index, digest (short), compressed size, command that created
the layer, and optionally the list of files added/modified/deleted.

## Global Flags

| Flag | Description |
|------|-------------|
| `--output json` | Emit JSON instead of human-readable output |
| `--local` | Only check local Docker daemon; error if not found |
| `--remote` | Only check remote registry; skip local daemon |
| `--debug` | Enable verbose logging for troubleshooting |

`--local` and `--remote` are mutually exclusive. Without either flag, the tool tries the local
daemon first and falls back to the remote registry.

## Image Resolution Flow

```
loader.Load(ref, source)
        │
        ├─ source=LocalOnly  → daemon transport only
        │                       error if not found
        │
        ├─ source=RemoteOnly → registry transport only
        │                       auth via ~/.docker/config.json keychain
        │
        └─ source=Auto       → try daemon first
                                fall back to registry
```

For remote images, only the manifest and config blob are fetched — layer tarballs are not
downloaded unless explicitly needed (and streamed rather than buffered when they are).

## Error Handling

- User-facing errors go to stderr; no stack traces by default
- `--debug` enables verbose logging
- Exit codes: `1` = usage error, `2` = image not found, `3` = registry/daemon unreachable

## Testing Strategy

- **Unit tests**: mock the `ggcr` `Image` interface; test output rendering with fixed fixtures
- **Integration tests**: use `ggcr`'s in-memory registry test package; no Docker daemon required in CI
- **Daemon tests**: tagged `//go:build integration`, skipped in normal `go test ./...`
