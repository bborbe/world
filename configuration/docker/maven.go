package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Maven struct {
	Image world.Image
}

func (p *Maven) Childs() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "maven",
			GitRepo:         "https://github.com/bborbe/maven.git",
			SourceDirectory: "github.com/bborbe/maven",
			Package:         "github.com/bborbe/maven",
			Image:           p.Image,
		}),
	}
}

func (p *Maven) Applier() world.Applier {
	return &docker.Uploader{
		Image: p.Image,
	}
}

func (p *Maven) Validate(ctx context.Context) error {
	if err := p.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
