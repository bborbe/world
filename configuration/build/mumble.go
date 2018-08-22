package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Mumble struct {
	Image docker.Image
}

func (m *Mumble) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/mumble.git",
			Image:     m.Image,
			GitBranch: docker.GitBranch(m.Image.Tag),
		}),
	}
}

func (m *Mumble) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
