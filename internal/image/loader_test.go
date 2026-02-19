package image_test

import (
	"errors"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
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

func TestLoader_LocalOnly(t *testing.T) {
	want := randomImage(t)
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return want, nil },
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("should not call remote") },
	)
	got, err := l.Load("alpine:latest", image.LocalOnly)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("expected daemon image to be returned")
	}
}

func TestLoader_LocalOnly_NotFound(t *testing.T) {
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("not found") },
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("should not call remote") },
	)
	_, err := l.Load("alpine:latest", image.LocalOnly)
	if err == nil {
		t.Error("expected error for missing local image")
	}
}

func TestLoader_RemoteOnly(t *testing.T) {
	want := randomImage(t)
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("should not call daemon") },
		func(_ name.Reference) (v1.Image, error) { return want, nil },
	)
	got, err := l.Load("alpine:latest", image.RemoteOnly)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("expected registry image to be returned")
	}
}

func TestLoader_Auto_PrefersDaemon(t *testing.T) {
	daemon := randomImage(t)
	registry := randomImage(t)
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return daemon, nil },
		func(_ name.Reference) (v1.Image, error) { return registry, nil },
	)
	got, err := l.Load("alpine:latest", image.Auto)
	if err != nil {
		t.Fatal(err)
	}
	if got != daemon {
		t.Error("expected daemon image to be preferred in Auto mode")
	}
}

func TestLoader_Auto_FallsBackToRemote(t *testing.T) {
	want := randomImage(t)
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return nil, errors.New("not in daemon") },
		func(_ name.Reference) (v1.Image, error) { return want, nil },
	)
	got, err := l.Load("alpine:latest", image.Auto)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("expected fallback to registry image")
	}
}

func TestLoader_InvalidRef(t *testing.T) {
	l := image.NewLoaderWithFetchers(
		func(_ name.Reference) (v1.Image, error) { return nil, nil },
		func(_ name.Reference) (v1.Image, error) { return nil, nil },
	)
	_, err := l.Load("not a valid::ref", image.Auto)
	if err == nil {
		t.Error("expected error for invalid image reference")
	}
}
