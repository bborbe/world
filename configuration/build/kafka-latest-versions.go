package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaLatestVersions struct {
	Image docker.Image
}

func (t *KafkaLatestVersions) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *KafkaLatestVersions) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-latest-versions",
				GitRepo:         "https://github.com/bborbe/kafka-latest-versions.git",
				SourceDirectory: "github.com/bborbe/kafka-latest-versions",
				Package:         "github.com/bborbe/kafka-latest-versions",
				Image:           i.Image,
			},
		},
	}
}

func (i *KafkaLatestVersions) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}