package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Monitoring struct {
	Image docker.Image
}

func (i *Monitoring) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "monitoring",
			GitRepo:         "https://github.com/bborbe/monitoring.git",
			SourceDirectory: "github.com/bborbe/monitoring",
			Package:         "github.com/bborbe/monitoring/bin/monitoring_server",
			Image:           i.Image,
		}),
	}
}

func (i *Monitoring) Applier() world.Applier {
	return &docker.Uploader{
		Image: i.Image,
	}
}

func (i *Monitoring) Validate(ctx context.Context) error {
	if err := i.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
