package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type HelloWorld struct {
	Image docker.Image
}

func (t *HelloWorld) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
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

func (h *HelloWorld) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: h.Image,
	}, nil
}
