package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Now struct {
	Image docker.Image
}

func (t *Now) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}
func (p *Now) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "now",
				GitRepo:         "https://github.com/bborbe/now.git",
				SourceDirectory: "github.com/bborbe/now",
				Package:         "github.com/bborbe/now/cmd/now-server",
				Image:           p.Image,
			},
		},
	}
}

func (p *Now) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}