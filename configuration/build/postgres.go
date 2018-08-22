package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Postgres struct {
	Image docker.Image
}

func (n *Postgres) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Registry:   "docker.io",
				Repository: "postgres",
				Tag:        n.Image.Tag,
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *Postgres) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
