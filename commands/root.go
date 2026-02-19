package commands

import (
	"github.com/spf13/cobra"
	"github.com/thisisnotashwin/imgutil/internal/format"
	"github.com/thisisnotashwin/imgutil/internal/image"
)

// GlobalFlags holds flags inherited by all subcommands.
type GlobalFlags struct {
	Output string
	Local  bool
	Remote bool
	Debug  bool
}

// NewRootCmd builds the root cobra command with all subcommands attached.
// loader is injected so tests can provide a mock-backed loader.
func NewRootCmd(loader *image.Loader) *cobra.Command {
	flags := &GlobalFlags{}

	root := &cobra.Command{
		Use:   "imgutil",
		Short: "Inspect Docker images from local daemon or remote registries",
	}

	root.PersistentFlags().StringVarP(&flags.Output, "output", "o", "human", `Output format: "human" or "json"`)
	root.PersistentFlags().BoolVar(&flags.Local, "local", false, "Only check local Docker daemon")
	root.PersistentFlags().BoolVar(&flags.Remote, "remote", false, "Only check remote registry")
	root.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug logging")
	root.MarkFlagsMutuallyExclusive("local", "remote")

	root.AddCommand(newInspectCmd(loader, flags))
	root.AddCommand(newLayersCmd(loader, flags))

	return root
}

func sourceFromFlags(flags *GlobalFlags) image.Source {
	if flags.Local {
		return image.LocalOnly
	}
	if flags.Remote {
		return image.RemoteOnly
	}
	return image.Auto
}

func formatFromFlags(flags *GlobalFlags) format.Format {
	if flags.Output == "json" {
		return format.JSON
	}
	return format.Human
}
