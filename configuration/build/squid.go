package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Squid struct {
	Image docker.Image
}

func (w *Squid) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}

func (p *Squid) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/squid.git",
				Image:     p.Image,
				GitBranch: docker.GitBranch(p.Image.Tag),
			},
		},
	}
}

func (p *Squid) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
