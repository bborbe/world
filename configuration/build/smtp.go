package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Smtp struct {
	Image docker.Image
}

func (m *Smtp) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/smtp.git",
			Image:     m.Image,
			GitBranch: docker.GitBranch(m.Image.Tag),
		}),
	}
}

func (m *Smtp) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
