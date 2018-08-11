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

func (o *Openldap) Childs() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/openldap.git",
			Image:     o.Image,
			GitBranch: world.GitBranch(o.Image.Tag),
		}),
	}
}

func (o *Openldap) Applier() world.Applier {
	return &docker.Uploader{
		Image: o.Image,
	}
}

func (o *Openldap) Validate(ctx context.Context) error {
	if err := o.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
