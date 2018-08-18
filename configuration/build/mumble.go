package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Mumble struct {
	Image docker.Image
}

func (m *Mumble) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/mumble.git",
			Image:     m.Image,
			GitBranch: docker.GitBranch(m.Image.Tag),
		}),
	}
}

func (m *Mumble) Applier() world.Applier {
	return &docker.Uploader{
		Image: m.Image,
	}
}

func (m *Mumble) Validate(ctx context.Context) error {
	if err := m.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
