---
name: security-review
description: Use when about to create a PR, push a branch, or run commit-push-pr. Runs the security-reviewer agent and blocks PR creation on any HIGH or CRITICAL findings.
---

# Security Review Gate

Security review is **mandatory** before any PR is created. The PR must not be created if the verdict is FAIL.

## Process

```dot
digraph security_gate {
    "About to create PR?" [shape=doublecircle];
    "Invoke security-reviewer agent" [shape=box];
    "Verdict PASS?" [shape=diamond];
    "Proceed: create PR" [shape=box];
    "STOP: show findings to user" [shape=box];
    "User fixes issues?" [shape=diamond];
    "Re-run security-reviewer" [shape=box];

    "About to create PR?" -> "Invoke security-reviewer agent";
    "Invoke security-reviewer agent" -> "Verdict PASS?";
    "Verdict PASS?" -> "Proceed: create PR" [label="yes"];
    "Verdict PASS?" -> "STOP: show findings to user" [label="no"];
    "STOP: show findings to user" -> "User fixes issues?";
    "User fixes issues?" -> "Re-run security-reviewer" [label="yes"];
    "User fixes issues?" -> "Abandon PR" [label="no"];
    "Re-run security-reviewer" -> "Verdict PASS?";
}
```

## Rules

- **NEVER** call `gh pr create` if the security-reviewer returned FAIL
- **NEVER** skip the review because "it's a small change" or "just a refactor"
- **NEVER** proceed if the security-reviewer agent produced an error or incomplete output — treat as FAIL
- The hook will also block `gh pr create` automatically, but Claude must enforce this before even reaching the hook

## Invoking the Agent

Use the Task tool with `subagent_type: security-reviewer`:

```
Task(
  subagent_type: "security-reviewer",
  description: "Security review before PR",
  prompt: "Review the current branch for security issues before PR creation. Run git diff main...HEAD, audit all changed Go files, run go vet and golangci-lint, and return your structured verdict."
)
```

## On FAIL

Show the user the full findings table. Do not create the PR. Say:

> "Security review returned FAIL. The following issues must be resolved before this can be submitted as a PR: [findings]. Please fix these and I will re-run the review."

## On PASS

Proceed normally with `commit-push-pr` or `gh pr create`. Include "Security review: PASS" as a note in the PR description.

## Common Rationalizations — All Invalid

| Excuse | Reality |
|--------|---------|
| "It's just docs/tests" | Changed test helpers can introduce security issues. Always review. |
| "golangci-lint already runs in CI" | CI is a backstop, not a substitute for pre-PR review. |
| "The change is trivial" | Trivial changes introduced Heartbleed. No exceptions. |
| "The user said to skip review" | Only the user explicitly saying "skip security review" allows skipping — not urgency or convenience. |
| "The hook will catch it" | Claude must also enforce this. Defense in depth. |
