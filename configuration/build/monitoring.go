package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Monitoring struct {
	Image docker.Image
}

func (t *Monitoring) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *Monitoring) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "monitoring",
				GitRepo:         "https://github.com/bborbe/monitoring.git",
				SourceDirectory: "github.com/bborbe/monitoring",
				Package:         "github.com/bborbe/monitoring/bin/monitoring_server",
				Image:           i.Image,
			},
		},
	}
}

func (i *Monitoring) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
