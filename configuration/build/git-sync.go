package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type GitSync struct {
	Image docker.Image
}

func (t *GitSync) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (m *GitSync) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/git-sync.git",
				Image:     m.Image,
				GitBranch: docker.GitBranch(m.Image.Tag),
			},
		},
	}
}

func (m *GitSync) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
