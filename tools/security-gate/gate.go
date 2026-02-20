package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// execFunc is a function that runs an external command and returns combined
// stdout+stderr output and the exit code. Injected so tests don't need real
// subprocesses.
type execFunc func(name string, args []string, dir string) (output string, exitCode int)

// Result is what Run returns to the caller (and ultimately to main).
type Result struct {
	Code   int    // 0 = pass, 1 = fail
	Output string // text to print to stdout
}

type toolInput struct {
	Command string `json:"command"`
}

type checkFailure struct {
	tool   string
	output string
}

// ansiPattern matches ANSI terminal escape sequences.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes ANSI escape sequences from s to prevent terminal
// manipulation via crafted source file content embedded in tool output.
func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// Run is the core gate logic. It parses the Claude tool input JSON, short-
// circuits if the command is not "gh pr create", otherwise runs three checks
// and returns a structured Result.
func Run(input, repoDir string, exec execFunc) Result {
	cmd, ok := parseCommand(input)
	if !ok {
		return Result{Code: 0}
	}
	if !strings.Contains(strings.ToLower(cmd), "gh pr create") {
		return Result{Code: 0}
	}
	failures := runChecks(repoDir, exec)
	return formatResult(failures)
}

func runChecks(repoDir string, exec execFunc) []checkFailure {
	var failures []checkFailure

	// go vet
	out, code := exec("go", []string{"vet", "./..."}, repoDir)
	if code != 0 {
		failures = append(failures, checkFailure{tool: "go vet", output: strings.TrimSpace(out)})
	}

	// golangci-lint
	out, code = exec("golangci-lint", []string{
		"run",
		"--enable=gosec,errcheck,bodyclose,noctx",
		"--timeout=120s",
	}, repoDir)
	if code != 0 {
		failures = append(failures, checkFailure{tool: "golangci-lint", output: strings.TrimSpace(out)})
	}

	// secret pattern grep
	out, code = exec("grep", []string{"-rn", "--include=*.go", "-E", secretPatterns(), "."}, repoDir)
	// grep exits 0 when it finds matches — matches are the problem here
	if code == 0 && strings.TrimSpace(out) != "" {
		var kept []string
		for _, line := range strings.Split(out, "\n") {
			if line == "" {
				continue
			}
			// Extract the file path (text before the first ':') so that
			// pattern content in the matched line cannot spoof the filter.
			path := line
			if i := strings.Index(line, ":"); i >= 0 {
				path = line[:i]
			}
			if strings.Contains(path, "vendor/") || strings.Contains(path, "_test.go") {
				continue
			}
			kept = append(kept, line)
		}
		if len(kept) > 0 {
			failures = append(failures, checkFailure{
				tool:   "secret/insecure pattern scan",
				output: strings.Join(kept, "\n"),
			})
		}
	}

	return failures
}

func formatResult(failures []checkFailure) Result {
	if len(failures) == 0 {
		return Result{
			Code:   0,
			Output: "Security gate: PASS — all checks clean. Proceeding with PR creation.",
		}
	}

	sep := strings.Repeat("=", 60)
	var sb strings.Builder
	sb.WriteString(sep + "\n")
	sb.WriteString("SECURITY GATE: FAIL — PR creation blocked\n")
	sb.WriteString(sep + "\n\n")
	sb.WriteString("The following automated security checks failed:\n\n")
	for _, f := range failures {
		sb.WriteString(fmt.Sprintf("[%s]\n", f.tool))
		sb.WriteString(stripANSI(f.output) + "\n\n")
	}
	sb.WriteString("Fix the issues above, then retry PR creation.\n")
	sb.WriteString(sep + "\n")

	return Result{Code: 1, Output: sb.String()}
}

func parseCommand(input string) (string, bool) {
	var ti toolInput
	if err := json.Unmarshal([]byte(input), &ti); err != nil {
		return "", false
	}
	return ti.Command, true
}

// secretPatterns returns the grep -E pattern string for secret scanning.
func secretPatterns() string {
	return strings.Join([]string{
		`(password|passwd|secret|api_key|apikey|token)\s*=\s*"[^"]+"`,
		`InsecureSkipVerify\s*:\s*true`,
		`math/rand.*Intn.*crypto`,
	}, "|")
}
