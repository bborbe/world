package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Smtp struct {
	Image docker.Image
}

func (w *Smtp) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}

func (m *Smtp) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/smtp.git",
				Image:     m.Image,
				GitBranch: docker.GitBranch(m.Image.Tag),
			},
		},
	}
}

func (m *Smtp) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
