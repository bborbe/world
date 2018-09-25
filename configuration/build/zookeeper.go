package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Zookeeper struct {
	Image docker.Image
}

func (k *Zookeeper) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *Zookeeper) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:         "https://github.com/kubernetes/contrib.git",
				Image:           k.Image,
				GitBranch:       "master",
				SourceDirectory: "statefulsets/zookeeper",
			},
		},
	}
}

func (k *Zookeeper) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
