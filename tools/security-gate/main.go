package main

import (
	"fmt"
	"os"
	"os/exec"
)

// allowedCommands is the explicit allowlist of binaries realExec may spawn.
// Any name not in this set is rejected, closing the G204 design-surface risk.
var allowedCommands = map[string]bool{
	"go":            true,
	"golangci-lint": true,
	"grep":          true,
}

func realExec(name string, args []string, dir string) (string, int) {
	if !allowedCommands[name] {
		return fmt.Sprintf("security-gate: blocked disallowed command %q", name), 1
	}
	cmd := exec.Command(name, args...) //nolint:gosec // G204: name is allowlist-validated above
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), exitErr.ExitCode()
		}
		return string(out), 1
	}
	return string(out), 0
}

func main() {
	input := os.Getenv("CLAUDE_TOOL_INPUT")
	if input == "" {
		input = "{}"
	}

	repoDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if repoDir == "" {
		var err error
		repoDir, err = os.Getwd()
		if err != nil {
			// Don't block all Bash usage on an env error.
			fmt.Fprintln(os.Stderr, "security-gate: could not determine working directory:", err)
			os.Exit(0)
		}
	}

	result := Run(input, repoDir, realExec)
	if result.Output != "" {
		fmt.Println(result.Output)
	}
	os.Exit(result.Code)
}
