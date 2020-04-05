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

type Redis struct {
	Image docker.Image
}

func (r *Redis) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.Image,
	)
}

func (r *Redis) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "redis",
					Tag:        r.Image.Tag,
				},
				TargetImage: r.Image,
			},
		),
	}
}

func (r *Redis) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: r.Image,
	}, nil
}
