package format

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// Format controls output rendering.
type Format string

const (
	Human Format = "human"
	JSON  Format = "json"
)

// InspectData holds image configuration metadata for output.
type InspectData struct {
	Reference  string            `json:"reference"`
	Digest     string            `json:"digest"`
	OS         string            `json:"os"`
	Arch       string            `json:"arch"`
	Created    string            `json:"created"`
	SizeBytes  int64             `json:"size_bytes"`
	Entrypoint []string          `json:"entrypoint"`
	Cmd        []string          `json:"cmd"`
	Env        []string          `json:"env"`
	Ports      []string          `json:"ports"`
	Labels     map[string]string `json:"labels"`
}

// LayerData holds per-layer information for output.
type LayerData struct {
	Index   int    `json:"index"`
	Digest  string `json:"digest"`
	Size    int64  `json:"size"`
	Command string `json:"command"`
}

// PrintInspect writes image metadata to w in the requested format.
func PrintInspect(w io.Writer, data InspectData, f Format) error {
	if f == JSON {
		return printJSON(w, data)
	}
	return printInspectHuman(w, data)
}

// PrintLayers writes layer data to w in the requested format.
func PrintLayers(w io.Writer, layers []LayerData, f Format) error {
	if f == JSON {
		return printJSON(w, layers)
	}
	return printLayersHuman(w, layers)
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printInspectHuman(w io.Writer, data InspectData) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "Reference:\t%s\n", data.Reference)
	_, _ = fmt.Fprintf(tw, "Digest:\t%s\n", data.Digest)
	_, _ = fmt.Fprintf(tw, "OS/Arch:\t%s/%s\n", data.OS, data.Arch)
	_, _ = fmt.Fprintf(tw, "Created:\t%s\n", data.Created)
	_, _ = fmt.Fprintf(tw, "Size:\t%s\n", HumanSize(data.SizeBytes))
	_, _ = fmt.Fprintf(tw, "Entrypoint:\t%v\n", data.Entrypoint)
	_, _ = fmt.Fprintf(tw, "Cmd:\t%v\n", data.Cmd)
	if len(data.Env) > 0 {
		_, _ = fmt.Fprintf(tw, "Env:\n")
		for _, e := range data.Env {
			_, _ = fmt.Fprintf(tw, "  %s\t\n", e)
		}
	}
	if len(data.Ports) > 0 {
		_, _ = fmt.Fprintf(tw, "Ports:\t%v\n", data.Ports)
	}
	if len(data.Labels) > 0 {
		_, _ = fmt.Fprintf(tw, "Labels:\n")
		for k, v := range data.Labels {
			_, _ = fmt.Fprintf(tw, "  %s=%s\t\n", k, v)
		}
	}
	return tw.Flush()
}

func printLayersHuman(w io.Writer, layers []LayerData) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "#\tDIGEST\tSIZE\tCOMMAND\n")
	for _, l := range layers {
		digest := l.Digest
		if len(digest) > 19 {
			digest = digest[:19]
		}
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", l.Index+1, digest, HumanSize(l.Size), l.Command)
	}
	return tw.Flush()
}

// HumanSize formats a byte count as a human-readable string (exported for testing).
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
