package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type AuthHttpProxy struct {
	Image docker.Image
}

func (t *AuthHttpProxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *AuthHttpProxy) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "auth-http-proxy",
				GitRepo:         "https://github.com/bborbe/auth-http-proxy.git",
				SourceDirectory: "github.com/bborbe/auth-http-proxy",
				Package:         "github.com/bborbe/auth-http-proxy",
				Image:           i.Image,
			},
		},
	}
}

func (i *AuthHttpProxy) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
