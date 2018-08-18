package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Now struct {
	Image docker.Image
}

func (p *Now) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "now",
			GitRepo:         "https://github.com/bborbe/now.git",
			SourceDirectory: "github.com/bborbe/now",
			Package:         "github.com/bborbe/now/cmd/now-server",
			Image:           p.Image,
		}),
	}
}

func (p *Now) Applier() world.Applier {
	return &docker.Uploader{
		Image: p.Image,
	}
}

func (p *Now) Validate(ctx context.Context) error {
	if err := p.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
