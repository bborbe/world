package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Postgres struct {
	Image docker.Image
}

func (n *Postgres) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Registry:   "docker.io",
				Repository: "postgres",
				Tag:        n.Image.Tag,
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *Postgres) Applier() world.Applier {
	return &docker.Uploader{
		Image: n.Image,
	}
}

func (n *Postgres) Validate(ctx context.Context) error {
	if err := n.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
