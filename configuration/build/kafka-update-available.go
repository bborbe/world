package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaUpdateAvailable struct {
	Image docker.Image
}

func (k *KafkaUpdateAvailable) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaUpdateAvailable) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-update-available",
				GitRepo:         "https://github.com/bborbe/kafka-update-available.git",
				SourceDirectory: "github.com/bborbe/kafka-update-available",
				Package:         "github.com/bborbe/kafka-update-available",
				Image:           k.Image,
			},
		},
	}
}

func (k *KafkaUpdateAvailable) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
