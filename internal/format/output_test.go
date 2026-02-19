package format_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thisisnotashwin/imgutil/internal/format"
)

func TestPrintInspect_JSON(t *testing.T) {
	data := format.InspectData{
		Reference: "alpine:latest",
		Digest:    "sha256:abc",
		OS:        "linux",
		Arch:      "amd64",
	}
	var buf bytes.Buffer
	if err := format.PrintInspect(&buf, data, format.JSON); err != nil {
		t.Fatal(err)
	}
	var got format.InspectData
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, buf.String())
	}
	if got.Reference != data.Reference {
		t.Errorf("got reference %q, want %q", got.Reference, data.Reference)
	}
	if got.OS != data.OS {
		t.Errorf("got OS %q, want %q", got.OS, data.OS)
	}
}

func TestPrintInspect_Human(t *testing.T) {
	data := format.InspectData{
		Reference: "alpine:latest",
		Digest:    "sha256:abc",
		OS:        "linux",
		Arch:      "amd64",
	}
	var buf bytes.Buffer
	if err := format.PrintInspect(&buf, data, format.Human); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "alpine:latest") {
		t.Error("human output missing image reference")
	}
	if !strings.Contains(out, "linux/amd64") {
		t.Error("human output missing OS/arch")
	}
}

func TestPrintLayers_JSON(t *testing.T) {
	layers := []format.LayerData{
		{Index: 0, Digest: "sha256:abc", Size: 1024, Command: "ADD file:..."},
	}
	var buf bytes.Buffer
	if err := format.PrintLayers(&buf, layers, format.JSON); err != nil {
		t.Fatal(err)
	}
	var got []format.LayerData
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}
	if len(got) != 1 {
		t.Errorf("got %d layers, want 1", len(got))
	}
	if got[0].Command != "ADD file:..." {
		t.Errorf("got command %q, want %q", got[0].Command, "ADD file:...")
	}
}

func TestPrintLayers_Human(t *testing.T) {
	layers := []format.LayerData{
		{Index: 0, Digest: "sha256:abcdef123456", Size: 7_000_000, Command: "ADD file:..."},
	}
	var buf bytes.Buffer
	if err := format.PrintLayers(&buf, layers, format.Human); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "DIGEST") {
		t.Error("human output missing DIGEST header")
	}
	if !strings.Contains(out, "ADD file:...") {
		t.Error("human output missing command")
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{512, "512 B"},
		{1024, "1.00 KB"},
		{7_340_032, "7.00 MB"},
	}
	for _, tc := range cases {
		got := format.HumanSize(tc.bytes)
		if got != tc.want {
			t.Errorf("HumanSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}
