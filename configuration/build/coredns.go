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

type CoreDns struct {
	Image docker.Image
}

func (c *CoreDns) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Image,
	)
}

func (c *CoreDns) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "coredns/coredns",
					Tag:        c.Image.Tag,
				},
				TargetImage: c.Image,
			},
		),
	}
}

func (c *CoreDns) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
