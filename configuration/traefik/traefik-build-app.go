// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traefik

import (
	"context"
	"github.com/bborbe/world/pkg/build"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BuildApp struct {
	Image docker.Image
}

func (b *BuildApp) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *BuildApp) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "traefik",
					Tag:        b.Image.Tag,
				},
				TargetImage: b.Image,
			},
		),
	}
}

func (b *BuildApp) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
