package commands_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/thisisnotashwin/imgutil/commands"
	"github.com/thisisnotashwin/imgutil/internal/image"
)

func randomImage(t *testing.T) v1.Image {
	t.Helper()
	img, err := random.Image(1024, 2)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func daemonLoader(img v1.Image) *image.Loader {
	return image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return img, nil },
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("no remote") },
	)
}

func TestInspectCmd_HumanOutput(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"inspect", "alpine:latest"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Reference") {
		t.Errorf("output missing Reference field\ngot: %s", out)
	}
	if !strings.Contains(out, "Digest") {
		t.Errorf("output missing Digest field\ngot: %s", out)
	}
}

func TestInspectCmd_JSONOutput(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"inspect", "--output", "json", "alpine:latest"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, buf.String())
	}
	if !strings.Contains(buf.String(), `"reference"`) {
		t.Errorf("JSON output missing reference field\ngot: %s", buf.String())
	}
}

func TestInspectCmd_RequiresArgument(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)
	root.SetArgs([]string{"inspect"})

	if err := root.Execute(); err == nil {
		t.Error("expected error when no image argument provided")
	}
}

func TestInspectCmd_MutuallyExclusiveFlags(t *testing.T) {
	loader := daemonLoader(randomImage(t))
	root := commands.NewRootCmd(loader)

	var buf bytes.Buffer
	root.SetErr(&buf)
	root.SetArgs([]string{"inspect", "--local", "--remote", "alpine:latest"})

	if err := root.Execute(); err == nil {
		t.Error("expected error when --local and --remote both set")
	}
}
