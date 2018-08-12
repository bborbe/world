package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Webdav struct {
	Image docker.Image
}

func (o *Webdav) Childs() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/webdav.git",
			Image:     o.Image,
			GitBranch: docker.GitBranch(o.Image.Tag),
		}),
	}
}

func (o *Webdav) Applier() world.Applier {
	return &docker.Uploader{
		Image: o.Image,
	}
}

func (o *Webdav) Validate(ctx context.Context) error {
	if err := o.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
