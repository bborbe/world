package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Password struct {
	Image docker.Image
}

func (t *Password) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (p *Password) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "password",
				GitRepo:         "https://github.com/bborbe/password.git",
				SourceDirectory: "github.com/bborbe/password",
				Package:         "github.com/bborbe/password/cmd/password-server",
				Image:           p.Image,
			},
		},
	}
}

func (p *Password) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
