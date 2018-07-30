package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Openldap struct {
	Image world.Image
}

func (m *Openldap) Childs() []world.Configuration {
	return []world.Configuration{
		&docker.Builder{
			GitRepo: "https://github.com/bborbe/openldap.git",
			Image:   m.Image,
		},
	}
}

func (m *Openldap) Applier() world.Applier {
	return &docker.Uploader{
		Image: m.Image,
	}
}

func (m *Openldap) Validate(ctx context.Context) error {
	if err := m.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
