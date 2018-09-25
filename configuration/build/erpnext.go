package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Erpnext struct {
	Image docker.Image
}

func (e *Erpnext) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.Image,
	)
}

func (e *Erpnext) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/erpnext.git",
				Image:     e.Image,
				GitBranch: docker.GitBranch(e.Image.Tag),
			},
		},
	}
}

func (e *Erpnext) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: e.Image,
	}, nil
}
