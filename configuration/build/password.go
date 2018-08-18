package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
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

func (p *Password) Applier() world.Applier {
	return &docker.Uploader{
		Image: p.Image,
	}
}

func (p *Password) Validate(ctx context.Context) error {
	if err := p.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
