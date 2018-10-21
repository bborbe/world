package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaK8sVersionCollector struct {
	Image docker.Image
}

func (t *KafkaK8sVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *KafkaK8sVersionCollector) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-k8s-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-k8s-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-k8s-version-collector",
				Package:         "github.com/bborbe/kafka-k8s-version-collector",
				Image:           i.Image,
			},
		},
	}
}

func (i *KafkaK8sVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
