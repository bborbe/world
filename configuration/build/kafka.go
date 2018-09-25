package build

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kafka struct {
	Image docker.Image
}

func (k *Kafka) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *Kafka) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:         "https://github.com/kubernetes/contrib.git",
				Image:           k.Image,
				GitBranch:       "master",
				SourceDirectory: "statefulsets/kafka",
				BuildArgs: map[string]string{
					"KAFKA_VERSION": k.Image.Tag.String(),
					"KAFKA_DIST":    fmt.Sprintf("kafka_2.12-%s", k.Image.Tag.String()),
				},
			},
		},
	}
}

func (k *Kafka) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
