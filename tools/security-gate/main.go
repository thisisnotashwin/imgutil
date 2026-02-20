package main

import (
	"fmt"
	"os"
	"os/exec"
)

func realExec(name string, args []string, dir string) (string, int) {
	cmd := exec.Command(name, args...)
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
