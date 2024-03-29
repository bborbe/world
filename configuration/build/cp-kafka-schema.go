// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/build"
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

func (c *CpKafkaSchemaRegistry) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "confluentinc/cp-schema-registry",
					Tag:        c.Image.Tag,
				},
				TargetImage: c.Image,
			},
		),
	}, nil
}

func (c *CpKafkaSchemaRegistry) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
