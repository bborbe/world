package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaSample struct {
	Image docker.Image
}

func (t *KafkaSample) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *KafkaSample) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "http_server",
				GitRepo:         "https://github.com/bborbe/sample_kafka.git",
				SourceDirectory: "github.com/bborbe/sample_kafka",
				Package:         "github.com/bborbe/sample_kafka/http_server",
				Image:           i.Image,
			},
		},
	}
}

func (i *KafkaSample) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
