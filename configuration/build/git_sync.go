package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type GitSync struct {
	Image docker.Image
}

func (m *GitSync) Childs() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/git-sync.git",
			Image:     m.Image,
			GitBranch: docker.GitBranch(m.Image.Tag),
		}),
	}
}

func (m *GitSync) Applier() world.Applier {
	return &docker.Uploader{
		Image: m.Image,
	}
}

func (m *GitSync) Validate(ctx context.Context) error {
	if err := m.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
