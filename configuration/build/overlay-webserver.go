package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
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

func (o *OverlayWebserver) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}
