package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaExporter struct {
	Image docker.Image
}

func (k *KafkaExporter) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaExporter) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "kafka-exporter",
				GitRepo:         "https://github.com/danielqsj/kafka_exporter.git",
				SourceDirectory: "github.com/danielqsj/kafka_exporter",
				Package:         "github.com/danielqsj/kafka_exporter",
				Image:           k.Image,
			},
		},
	}
}

func (k *KafkaExporter) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
