package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KubeStateMetrics struct {
	Image docker.Image
}

func (t *KubeStateMetrics) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *KubeStateMetrics) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/coreos/kube-state-metrics",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *KubeStateMetrics) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}