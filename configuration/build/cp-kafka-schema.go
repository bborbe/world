package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type CpKafkaSchemaRegistry struct {
	Image docker.Image
}

func (c *CpKafkaSchemaRegistry) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Image,
	)
}

func (c *CpKafkaSchemaRegistry) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "confluentinc/cp-schema-registry",
					Tag:        c.Image.Tag,
				},
				TargetImage: c.Image,
			},
		},
	}
}

func (c *CpKafkaSchemaRegistry) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}