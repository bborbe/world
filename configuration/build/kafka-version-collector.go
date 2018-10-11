package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaVersionCollector struct {
	Image docker.Image
}

func (t *KafkaVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *KafkaVersionCollector) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-version-collector",
				Package:         "github.com/bborbe/kafka-version-collector",
				Image:           i.Image,
			},
		},
	}
}

func (i *KafkaVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
