package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Privoxy struct {
	Image docker.Image
}

func (p *Privoxy) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/privoxy.git",
			Image:     p.Image,
			GitBranch: docker.GitBranch(p.Image.Tag),
		}),
	}
}

func (p *Privoxy) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
