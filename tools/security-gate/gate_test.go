package main

import (
	"strings"
	"testing"
)

// fakeExec returns a function that acts as an execFunc. For each invocation
// it pops the next (output, code) pair from the provided list. Panics if
// called more times than pairs provided (catches unexpected calls).
func fakeExec(pairs []struct {
	out  string
	code int
}) execFunc {
	i := 0
	return func(name string, args []string, dir string) (string, int) {
		if i >= len(pairs) {
			panic("fakeExec called more times than expected")
		}
		p := pairs[i]
		i++
		return p.out, p.code
	}
}

func TestNonPRCommand(t *testing.T) {
	input := `{"command": "git status"}`
	result := Run(input, "/repo", fakeExec(nil))
	if result.Code != 0 {
		t.Errorf("expected code 0 for non-PR command, got %d", result.Code)
	}
	if result.Output != "" {
		t.Errorf("expected empty output for non-PR command, got %q", result.Output)
	}
}

func TestInvalidJSON(t *testing.T) {
	result := Run("not json", "/repo", fakeExec(nil))
	if result.Code != 0 {
		t.Errorf("expected code 0 for invalid JSON, got %d", result.Code)
	}
}

func TestMissingCommandField(t *testing.T) {
	result := Run(`{"other": "field"}`, "/repo", fakeExec(nil))
	if result.Code != 0 {
		t.Errorf("expected code 0 for missing command field, got %d", result.Code)
	}
}

func TestAllChecksPassing(t *testing.T) {
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0},          // go vet passes
		{"0 issues.", 0}, // golangci-lint passes
		{"", 1},          // grep: no matches (exit 1 = no match found)
	})
	result := Run(`{"command": "gh pr create --title foo"}`, "/repo", exec)
	if result.Code != 0 {
		t.Errorf("expected PASS, got code %d: %s", result.Code, result.Output)
	}
	if !strings.Contains(result.Output, "PASS") {
		t.Errorf("expected PASS in output, got: %s", result.Output)
	}
}

func TestGoVetFail(t *testing.T) {
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"./cmd/foo.go:12:5: unreachable code", 1}, // go vet fails
		{"0 issues.", 0},                            // golangci-lint passes
		{"", 1},                                     // grep: no matches
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL, got code %d", result.Code)
	}
	if !strings.Contains(result.Output, "go vet") {
		t.Errorf("expected 'go vet' in output, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", result.Output)
	}
}

func TestGolangciLintFail(t *testing.T) {
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0},                                  // go vet passes
		{"commands/foo.go:5:2: G104 (gosec)", 1}, // lint fails
		{"", 1},                                  // grep: no matches
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL, got code %d", result.Code)
	}
	if !strings.Contains(result.Output, "golangci-lint") {
		t.Errorf("expected 'golangci-lint' in output, got: %s", result.Output)
	}
}

func TestSecretPatternMatch(t *testing.T) {
	grepOut := "commands/foo.go:10:  password = \"hunter2\"\n"
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0},      // go vet passes
		{"", 0},      // golangci-lint passes
		{grepOut, 0}, // grep finds a match
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL on secret match, got code %d", result.Code)
	}
	if !strings.Contains(result.Output, "secret/insecure pattern scan") {
		t.Errorf("expected pattern scan label in output, got: %s", result.Output)
	}
}

func TestSecretPatternVendorExcluded(t *testing.T) {
	// All grep matches are in vendor or test files — should not fail.
	grepOut := "vendor/lib/foo.go:5: password = \"x\"\ncommands/foo_test.go:8: token = \"y\"\n"
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0},      // go vet passes
		{"", 0},      // golangci-lint passes
		{grepOut, 0}, // grep finds matches, but all filtered
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 0 {
		t.Errorf("expected PASS when only vendor/test matches, got code %d: %s", result.Code, result.Output)
	}
}

func TestUppercasePRCommandTriggersGate(t *testing.T) {
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0}, // go vet passes
		{"", 0}, // golangci-lint passes
		{"", 1}, // grep: no matches
	})
	result := Run(`{"command": "GH PR CREATE --title foo"}`, "/repo", exec)
	if result.Code != 0 {
		t.Errorf("expected PASS, got code %d: %s", result.Code, result.Output)
	}
	if !strings.Contains(result.Output, "PASS") {
		t.Errorf("expected PASS in output, got: %s", result.Output)
	}
}

func TestVendorFilterByPathNotContent(t *testing.T) {
	// Match line whose content contains "vendor/" but whose file path does not.
	// Old filter would silently suppress this; new path-based filter must keep it.
	grepOut := "commands/real.go:12: token = \"x\" // from vendor/lib\n"
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"", 0},      // go vet passes
		{"", 0},      // golangci-lint passes
		{grepOut, 0}, // grep finds match in non-vendor file
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL when match content contains 'vendor/' but path does not, got code %d", result.Code)
	}
}

func TestANSIStrippedFromOutput(t *testing.T) {
	ansiVetOut := "\x1b[31merror: bad code\x1b[0m"
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{ansiVetOut, 1}, // go vet fails with ANSI-coloured output
		{"", 0},         // golangci-lint passes
		{"", 1},         // grep: no matches
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL, got code %d", result.Code)
	}
	if strings.Contains(result.Output, "\x1b[") {
		t.Errorf("expected ANSI sequences stripped from output, got: %q", result.Output)
	}
	if !strings.Contains(result.Output, "error: bad code") {
		t.Errorf("expected plain text content preserved after stripping, got: %s", result.Output)
	}
}

func TestMultipleFailures(t *testing.T) {
	grepOut := "commands/foo.go:1: InsecureSkipVerify: true\n"
	exec := fakeExec([]struct {
		out  string
		code int
	}{
		{"vet error", 1},  // go vet fails
		{"lint error", 1}, // golangci-lint fails
		{grepOut, 0},      // grep finds match
	})
	result := Run(`{"command": "gh pr create"}`, "/repo", exec)
	if result.Code != 1 {
		t.Errorf("expected FAIL, got code %d", result.Code)
	}
	// All three tools should appear in output
	for _, tool := range []string{"go vet", "golangci-lint", "secret/insecure pattern scan"} {
		if !strings.Contains(result.Output, tool) {
			t.Errorf("expected %q in output, got: %s", tool, result.Output)
		}
	}
}
