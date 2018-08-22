package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type HelloWorld struct {
	Image docker.Image
}

func (h *HelloWorld) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/hello-world.git",
			Image:     h.Image,
			GitBranch: docker.GitBranch(h.Image.Tag),
		}),
	}
}

func (h *HelloWorld) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: h.Image,
	}, nil
}
