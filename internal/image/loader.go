package image

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Source controls where the loader looks for images.
type Source int

const (
	Auto      Source = iota // try daemon first, fall back to registry
	LocalOnly               // daemon only
	RemoteOnly              // registry only
)

// Loader resolves Docker image references to v1.Image values.
type Loader struct {
	fromDaemon   func(name.Reference) (v1.Image, error)
	fromRegistry func(name.Reference) (v1.Image, error)
}

// NewLoader returns a Loader backed by the local Docker daemon and the default
// remote registry keychain (~/.docker/config.json).
func NewLoader() *Loader {
	return &Loader{
		fromDaemon: func(ref name.Reference) (v1.Image, error) {
			return daemon.Image(ref)
		},
		fromRegistry: func(ref name.Reference) (v1.Image, error) {
			return remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		},
	}
}

// NewLoaderWithFetchers constructs a Loader with injected fetchers, for testing.
func NewLoaderWithFetchers(
	fromDaemon func(name.Reference) (v1.Image, error),
	fromRegistry func(name.Reference) (v1.Image, error),
) *Loader {
	return &Loader{fromDaemon: fromDaemon, fromRegistry: fromRegistry}
}

// Load resolves rawRef to a v1.Image using the given source strategy.
func (l *Loader) Load(rawRef string, src Source) (v1.Image, error) {
	ref, err := name.ParseReference(rawRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference %q: %w", rawRef, err)
	}

	switch src {
	case LocalOnly:
		img, err := l.fromDaemon(ref)
		if err != nil {
			return nil, fmt.Errorf("image %q not found in local daemon: %w", rawRef, err)
		}
		return img, nil

	case RemoteOnly:
		img, err := l.fromRegistry(ref)
		if err != nil {
			return nil, fmt.Errorf("image %q not found in remote registry: %w", rawRef, err)
		}
		return img, nil

	default: // Auto
		img, err := l.fromDaemon(ref)
		if err == nil {
			return img, nil
		}
		img, err = l.fromRegistry(ref)
		if err != nil {
			return nil, fmt.Errorf("image %q not found locally or in remote registry: %w", rawRef, err)
		}
		return img, nil
	}
}
