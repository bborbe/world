package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Postgres struct {
	Image docker.Image
}

func (t *Postgres) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *Postgres) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Repository: "postgres",
				Tag:        n.Image.Tag,
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *Postgres) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
