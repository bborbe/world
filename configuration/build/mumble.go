package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Mumble struct {
	Image docker.Image
}

func (t *Mumble) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (m *Mumble) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/mumble.git",
				Image:     m.Image,
				GitBranch: docker.GitBranch(m.Image.Tag),
			},
		},
	}
}

func (m *Mumble) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
