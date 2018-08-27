package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Traefik struct {
	Image docker.Image
}

func (w *Traefik) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}

func (n *Traefik) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "traefik",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *Traefik) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
