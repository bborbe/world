package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaAtlassianVersionCollector struct {
	Image docker.Image
}

func (t *KafkaAtlassianVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *KafkaAtlassianVersionCollector) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-atlassian-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-atlassian-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-atlassian-version-collector",
				Package:         "github.com/bborbe/kafka-atlassian-version-collector",
				Image:           i.Image,
			},
		},
	}
}

func (i *KafkaAtlassianVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
