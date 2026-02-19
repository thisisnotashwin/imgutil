---
name: new-command
description: Scaffold a new imgutil cobra subcommand following the project's established pattern
---

Create a new subcommand for imgutil. The command name is provided as an argument (e.g. `/new-command diff`).

## Pattern

Every command in this project follows this exact structure:

```go
package commands

import (
    "github.com/spf13/cobra"
    "github.com/thisisnotashwin/imgutil/internal/format"
    "github.com/thisisnotashwin/imgutil/internal/image"
)

func new<Name>Cmd(loader *image.Loader, flags *GlobalFlags) *cobra.Command {
    return &cobra.Command{
        Use:   "<name> <image>",
        Short: "<one-line description>",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            img, err := loader.Load(args[0], sourceFromFlags(flags))
            if err != nil {
                return err
            }

            // ... command-specific logic ...

            return format.Print<Name>(cmd.OutOrStdout(), data, formatFromFlags(flags))
        },
    }
}
```

## Steps

1. Create `commands/<name>.go` using the pattern above
2. Register it in `commands/root.go` alongside the existing `inspect` and `layers` commands:
   ```go
   rootCmd.AddCommand(newInspectCmd(loader, flags))
   rootCmd.AddCommand(newLayersCmd(loader, flags))
   rootCmd.AddCommand(new<Name>Cmd(loader, flags))  // add here
   ```
3. If new output types are needed, add the corresponding `Print<Name>` function and data struct in `internal/format/output.go`
4. Create `commands/<name>_test.go` mirroring the structure of `commands/inspect_test.go`

## Key conventions

- All errors wrapped with `fmt.Errorf("reading <thing>: %w", err)`
- Output always goes to `cmd.OutOrStdout()` (not `os.Stdout`) so tests can capture it
- Source selection always via `sourceFromFlags(flags)` — never hard-coded
- Format selection always via `formatFromFlags(flags)`
- No direct fmt.Print calls in command files — all rendering lives in `internal/format`
