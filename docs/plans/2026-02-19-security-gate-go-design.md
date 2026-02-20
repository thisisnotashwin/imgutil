# Security Gate — Go Rewrite Design

**Date:** 2026-02-19
**Status:** Approved

## Context

The existing `.claude/hooks/security_gate.py` intercepts `gh pr create` Bash calls and runs automated security checks (`go vet`, `golangci-lint`, secret pattern grep). The goal is to rewrite it in Go with unit tests, built as a pre-built binary invoked by the hook.

## Decision

**Binary placement:** `tools/security-gate/` (Go tool pattern with `//go:build ignore`)
**Build:** Pre-built via `make build-tools` → `bin/security-gate`
**Hook invocation:** `${CLAUDE_PROJECT_DIR}/bin/security-gate` (direct binary call)

## File Layout

```
tools/security-gate/
  main.go        – reads CLAUDE_TOOL_INPUT env var, calls Run(), prints output, exits 0/1
  gate.go        – Run(input, dir string, exec execFunc) Result; all checker logic
  gate_test.go   – unit tests with fake execFunc injected
bin/
  security-gate  – gitignored, built by make build-tools
docs/plans/
  2026-02-19-security-gate-go-design.md  – this file
```

## Architecture

### main.go

Thin entrypoint:

1. Read `CLAUDE_TOOL_INPUT` from env
2. Read `CLAUDE_PROJECT_DIR` from env (fallback: `os.Getwd()`)
3. Call `Run(input, dir, realExec)`
4. Print result output to stdout
5. `os.Exit(result.Code)`

### gate.go

Core logic, fully unit-testable via injected executor:

```go
type execFunc func(name string, args []string, dir string) (output string, exitCode int)

type Result struct {
    Code     int    // 0 = pass, 1 = fail
    Output   string // formatted text for stdout
}

func Run(toolInput, repoDir string, exec execFunc) Result
```

**Fast path:** if `"gh pr create"` not in parsed `.command` field → return `Result{Code: 0}` immediately.

**Checks (in order):**

1. `go vet ./...` — fail if exit code != 0
2. `golangci-lint run --enable=gosec,errcheck,bodyclose,noctx --out-format=line-number --timeout=120s` — fail if exit code != 0
3. `grep -rn --include=*.go -E <patterns> .` filtered to exclude `/vendor/` and `_test.go` lines — fail if any non-excluded matches found

Secret patterns (same as Python version):
- `(password|passwd|secret|api_key|apikey|token)\s*=\s*"[^"]+"`
- `InsecureSkipVerify\s*:\s*true`
- `math/rand.*Intn.*crypto`

**Output format** mirrors the Python version:
- Pass: single-line `Security gate: PASS — all checks clean. Proceeding with PR creation.`
- Fail: structured block with `SECURITY GATE: FAIL` header, per-tool findings, fix instructions

### gate_test.go

Unit tests using fake `execFunc` — no real subprocess calls:

| Test | Scenario |
|------|----------|
| `TestNonPRCommand` | command doesn't contain `gh pr create` → exit 0, no checks run |
| `TestAllChecksPassing` | all three checks return exit 0 / no matches → PASS |
| `TestGoVetFail` | go vet returns exit 1 → FAIL with go vet output |
| `TestGolangciLintFail` | lint returns exit 1 → FAIL with lint output |
| `TestSecretPatternMatch` | grep returns exit 0 with matches → FAIL (excluding vendor/test lines) |
| `TestSecretPatternVendorExcluded` | grep matches only in vendor/ → PASS |
| `TestInvalidJSON` | malformed CLAUDE_TOOL_INPUT → exit 0 (don't block on bad input) |
| `TestMissingCommandField` | JSON with no `command` key → exit 0 |

## Makefile Addition

```makefile
.PHONY: build-tools
build-tools: ## Build dev tools → ./bin/security-gate
    @mkdir -p bin
    @go build -o bin/security-gate ./tools/security-gate
```

## settings.json Hook Update

Replace the `python3 ...` command with:

```json
{
  "type": "command",
  "command": "\"${CLAUDE_PROJECT_DIR}/bin/security-gate\""
}
```

## Migration

1. Build and verify the Go binary produces identical output to the Python script
2. Update `settings.json` to invoke the binary
3. Delete `security_gate.py`
4. Add a note to README/CLAUDE.md: run `make build-tools` after cloning

## Trade-offs Considered

| Option | Decision |
|--------|----------|
| Flat package vs sub-package | Flat — tool is ~200 lines, no benefit to extra package boundary |
| Interface-based checkers | Rejected — injected `execFunc` gives sufficient testability at lower complexity |
| Build-on-demand in hook | Rejected — adds ~1s latency to every Bash tool call |
| Checked-in binary | Rejected — binaries in git are bad practice; `make build-tools` is cheap |
