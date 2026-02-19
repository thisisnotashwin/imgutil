package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/thisisnotashwin/imgutil/internal/format"
	"github.com/thisisnotashwin/imgutil/internal/image"
)

func newInspectCmd(loader *image.Loader, flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <image>",
		Short: "Display image configuration metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			img, err := loader.Load(args[0], sourceFromFlags(flags))
			if err != nil {
				return err
			}

			cfg, err := img.ConfigFile()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			digest, err := img.Digest()
			if err != nil {
				return fmt.Errorf("reading digest: %w", err)
			}

			size, err := img.Size()
			if err != nil {
				return fmt.Errorf("reading size: %w", err)
			}

			ports := make([]string, 0, len(cfg.Config.ExposedPorts))
			for p := range cfg.Config.ExposedPorts {
				ports = append(ports, string(p))
			}
			sort.Strings(ports)

			data := format.InspectData{
				Reference:  args[0],
				Digest:     digest.String(),
				OS:         cfg.OS,
				Arch:       cfg.Architecture,
				Created:    cfg.Created.UTC().Format("2006-01-02 15:04:05 UTC"),
				SizeBytes:  size,
				Entrypoint: cfg.Config.Entrypoint,
				Cmd:        cfg.Config.Cmd,
				Env:        cfg.Config.Env,
				Ports:      ports,
				Labels:     cfg.Config.Labels,
			}

			return format.PrintInspect(cmd.OutOrStdout(), data, formatFromFlags(flags))
		},
	}
}
