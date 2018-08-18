package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type OverlayWebserver struct {
	Image docker.Image
}

func (o *OverlayWebserver) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "overlay-server",
			GitRepo:         "https://github.com/bborbe/server.git",
			SourceDirectory: "github.com/bborbe/server",
			Package:         "github.com/bborbe/server/cmd/overlay-server",
			Image:           o.Image,
		}),
	}
}

func (o *OverlayWebserver) Applier() world.Applier {
	return &docker.Uploader{
		Image: o.Image,
	}
}

func (o *OverlayWebserver) Validate(ctx context.Context) error {
	if err := o.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
