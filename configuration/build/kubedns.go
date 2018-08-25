package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Kubedns struct {
	Image docker.Image
}

func (t *Kubedns) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *Kubedns) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Repository: "gcr.io/google_containers/kubedns-amd64",
				Tag:        n.Image.Tag,
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *Kubedns) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
