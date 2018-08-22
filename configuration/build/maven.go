package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Maven struct {
	Image docker.Image
}

func (p *Maven) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "maven",
			GitRepo:         "https://github.com/bborbe/maven.git",
			SourceDirectory: "github.com/bborbe/maven",
			Package:         "github.com/bborbe/maven",
			Image:           p.Image,
		}),
	}
}

func (p *Maven) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
