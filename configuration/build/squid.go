package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Squid struct {
	Image docker.Image
}

func (p *Squid) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/squid.git",
			Image:     p.Image,
			GitBranch: docker.GitBranch(p.Image.Tag),
		}),
	}
}

func (p *Squid) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
