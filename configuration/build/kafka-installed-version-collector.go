package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaInstalledVersionCollector struct {
	Image docker.Image
}

func (k *KafkaInstalledVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaInstalledVersionCollector) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-installed-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-installed-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-installed-version-collector",
				Package:         "github.com/bborbe/kafka-installed-version-collector",
				Image:           k.Image,
			},
		},
	}
}

func (k *KafkaInstalledVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}