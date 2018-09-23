package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type FrappeBench struct {
	Image docker.Image
}

func (t *FrappeBench) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (m *FrappeBench) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/frappe/frappe_docker.git",
				Image:     m.Image,
				GitBranch: docker.GitBranch(m.Image.Tag),
			},
		},
	}
}

func (m *FrappeBench) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
