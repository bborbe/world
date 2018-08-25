package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Webdav struct {
	Image docker.Image
}

func (w *Webdav) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}
func (o *Webdav) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/webdav.git",
			Image:     o.Image,
			GitBranch: docker.GitBranch(o.Image.Tag),
		}),
	}
}

func (o *Webdav) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}
