package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Traefik struct {
	Image docker.Image
}

func (n *Traefik) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Registry:   "docker.io",
				Repository: "traefik",
				Tag:        n.Image.Tag,
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *Traefik) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
