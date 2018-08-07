package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/configuration"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Mumble struct {
	Image world.Image
}

func (m *Mumble) Childs() []world.Configuration {
	return []world.Configuration{
		configuration.New().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/mumble.git",
			Image:     m.Image,
			GitBranch: world.GitBranch(m.Image.Tag),
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
		return errors.Wrap(err, "image missing")
	}
	return nil
}
