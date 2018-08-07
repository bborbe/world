package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/configuration"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type HelloWorld struct {
	Image world.Image
}

func (h *HelloWorld) Childs() []world.Configuration {
	return []world.Configuration{
		configuration.New().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/hello-world.git",
			Image:     h.Image,
			GitBranch: world.GitBranch(h.Image.Tag),
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
		return errors.Wrap(err, "image missing")
	}
	return nil
}
