package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Teamvault struct {
	Image docker.Image
}

func (p *Teamvault) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo: "https://github.com/bborbe/teamvault.git",
			Image:   p.Image,
			BuildArgs: docker.BuildArgs{
				"VERSION": p.Image.Tag.String(),
			},
			GitBranch: "master",
		}),
	}
}

func (p *Teamvault) Applier() world.Applier {
	return &docker.Uploader{
		Image: p.Image,
	}
}

func (p *Teamvault) Validate(ctx context.Context) error {
	if err := p.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
