package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Smtp struct {
	Image docker.Image
}

func (m *Smtp) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/smtp.git",
			Image:     m.Image,
			GitBranch: docker.GitBranch(m.Image.Tag),
		}),
	}
}

func (m *Smtp) Applier() world.Applier {
	return &docker.Uploader{
		Image: m.Image,
	}
}

func (m *Smtp) Validate(ctx context.Context) error {
	if err := m.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
