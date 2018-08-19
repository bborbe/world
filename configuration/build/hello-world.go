package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type HelloWorld struct {
	Image docker.Image
}

func (h *HelloWorld) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/hello-world.git",
			Image:     h.Image,
			GitBranch: docker.GitBranch(h.Image.Tag),
		}),
	}
}

func (h *HelloWorld) Applier() world.Applier {
	return &docker.Uploader{
		Image: h.Image,
	}
}

func (h *HelloWorld) Validate(ctx context.Context) error {
	if err := h.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
