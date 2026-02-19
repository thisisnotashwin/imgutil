package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thisisnotashwin/imgutil/commands"
)

func TestLayersCmd_HumanOutput(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"layers", "alpine:latest"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "DIGEST") {
		t.Errorf("output missing DIGEST header\ngot: %s", out)
	}
	if !strings.Contains(out, "SIZE") {
		t.Errorf("output missing SIZE header\ngot: %s", out)
	}
}

func TestLayersCmd_JSONOutput(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"layers", "--output", "json", "alpine:latest"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, buf.String())
	}
	if !strings.Contains(buf.String(), `"digest"`) {
		t.Errorf("JSON output missing digest field\ngot: %s", buf.String())
	}
}

func TestLayersCmd_RequiresArgument(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)
	root.SetArgs([]string{"layers"})

	if err := root.Execute(); err == nil {
		t.Error("expected error when no image argument provided")
	}
}
