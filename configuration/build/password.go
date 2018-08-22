package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Password struct {
	Image docker.Image
}

func (p *Password) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "password",
			GitRepo:         "https://github.com/bborbe/password.git",
			SourceDirectory: "github.com/bborbe/password",
			Package:         "github.com/bborbe/password/cmd/password-server",
			Image:           p.Image,
		}),
	}
}

func (p *Password) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
