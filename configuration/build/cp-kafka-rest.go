package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type CpKafkaRest struct {
	Image docker.Image
}

func (c *CpKafkaRest) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Image,
	)
}

func (c *CpKafkaRest) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "confluentinc/cp-kafka-rest",
					Tag:        c.Image.Tag,
				},
				TargetImage: c.Image,
			},
		},
	}
}

func (c *CpKafkaRest) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
