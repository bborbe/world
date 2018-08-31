package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type PrometheusAlertmanager struct {
	Image docker.Image
}

func (t *PrometheusAlertmanager) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *PrometheusAlertmanager) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/prometheus/alertmanager",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *PrometheusAlertmanager) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
