package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type OverlayWebserver struct {
	Image docker.Image
}

func (t *OverlayWebserver) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
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

func (o *OverlayWebserver) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}
