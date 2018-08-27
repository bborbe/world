package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Prometheus struct {
	Image docker.Image
}

func (t *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *Prometheus) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/prometheus/prometheus",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *Prometheus) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
