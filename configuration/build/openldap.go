package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Openldap struct {
	Image docker.Image
}

func (o *Openldap) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/openldap.git",
			Image:     o.Image,
			GitBranch: docker.GitBranch(o.Image.Tag),
		}),
	}
}

func (o *Openldap) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}

func (d *Openldap) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Image,
	)
}
