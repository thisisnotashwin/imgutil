#!/usr/bin/env python3
"""
Security gate hook for imgutil.
Intercepts `gh pr create` calls and runs automated security checks.
Exits non-zero (blocking the tool) if issues are found.
"""

import json
import os
import subprocess
import sys


def run(cmd, cwd):
    result = subprocess.run(
        cmd, cwd=cwd, capture_output=True, text=True
    )
    return result.returncode, result.stdout + result.stderr


def main():
    tool_input = os.environ.get("CLAUDE_TOOL_INPUT", "{}")
    try:
        data = json.loads(tool_input)
    except json.JSONDecodeError:
        sys.exit(0)

    command = data.get("command", "")
    if "gh pr create" not in command:
        sys.exit(0)

    repo_root = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    failures = []

    # --- go vet ---
    code, out = run(["go", "vet", "./..."], cwd=repo_root)
    if code != 0:
        failures.append(("go vet", out.strip()))

    # --- golangci-lint (gosec + errcheck + bodyclose) ---
    lint_cmd = [
        "golangci-lint", "run",
        "--enable=gosec,errcheck,bodyclose,noctx",
        "--out-format=line-number",
        "--timeout=120s",
    ]
    code, out = run(lint_cmd, cwd=repo_root)
    if code != 0:
        failures.append(("golangci-lint", out.strip()))

    # --- grep for obvious secrets ---
    secret_patterns = [
        r'(password|passwd|secret|api_key|apikey|token)\s*=\s*"[^"]+"',
        r'InsecureSkipVerify\s*:\s*true',
        r'math/rand.*Intn.*crypto',
    ]
    grep_cmd = ["grep", "-rn", "--include=*.go", "-E",
                "|".join(secret_patterns), "."]
    code, out = run(grep_cmd, cwd=repo_root)
    # grep exits 0 if matches found (which is bad here)
    if code == 0 and out.strip():
        # Filter out vendor and test files for grep
        lines = [l for l in out.splitlines()
                 if "/vendor/" not in l and "_test.go" not in l]
        if lines:
            failures.append(("secret/insecure pattern scan",
                             "\n".join(lines)))

    if not failures:
        print("Security gate: PASS — all checks clean. Proceeding with PR creation.")
        sys.exit(0)

    print("=" * 60)
    print("SECURITY GATE: FAIL — PR creation blocked")
    print("=" * 60)
    print()
    print("The following automated security checks failed:")
    print()
    for tool, output in failures:
        print(f"[{tool}]")
        print(output)
        print()
    print("Fix the issues above, then retry PR creation.")
    print("=" * 60)
    sys.exit(1)


if __name__ == "__main__":
    main()
