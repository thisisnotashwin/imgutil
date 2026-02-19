package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thisisnotashwin/imgutil/internal/format"
	"github.com/thisisnotashwin/imgutil/internal/image"
)

func newLayersCmd(loader *image.Loader, flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "layers <image>",
		Short: "Display per-layer breakdown of an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			img, err := loader.Load(args[0], sourceFromFlags(flags))
			if err != nil {
				return err
			}

			layers, err := img.Layers()
			if err != nil {
				return fmt.Errorf("reading layers: %w", err)
			}

			cfg, err := img.ConfigFile()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			layerData := make([]format.LayerData, 0, len(layers))
			for i, l := range layers {
				digest, err := l.Digest()
				if err != nil {
					return fmt.Errorf("reading layer %d digest: %w", i, err)
				}

				size, err := l.Size()
				if err != nil {
					return fmt.Errorf("reading layer %d size: %w", i, err)
				}

				createdBy := ""
				if i < len(cfg.History) {
					createdBy = cfg.History[i].CreatedBy
					createdBy = strings.TrimPrefix(createdBy, "|0 /bin/sh -c ")
					createdBy = strings.TrimPrefix(createdBy, "/bin/sh -c ")
					if len(createdBy) > 80 {
						createdBy = createdBy[:77] + "..."
					}
				}

				layerData = append(layerData, format.LayerData{
					Index:   i,
					Digest:  digest.String(),
					Size:    size,
					Command: createdBy,
				})
			}

			return format.PrintLayers(cmd.OutOrStdout(), layerData, formatFromFlags(flags))
		},
	}
}
