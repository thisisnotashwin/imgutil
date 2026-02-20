package main

import (
	"encoding/json"
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

// Run is the core gate logic. It parses the Claude tool input JSON, short-
// circuits if the command is not "gh pr create", otherwise runs three checks
// and returns a structured Result.
func Run(input, repoDir string, exec execFunc) Result {
	return Result{}
}

func runChecks(repoDir string, exec execFunc) []checkFailure {
	return nil
}

func formatResult(failures []checkFailure) Result {
	return Result{}
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
