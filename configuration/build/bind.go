package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Bind struct {
	Image docker.Image
}

func (o *Bind) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/bind.git",
			Image:     o.Image,
			GitBranch: docker.GitBranch(o.Image.Tag),
		}),
	}
}

func (o *Bind) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}
