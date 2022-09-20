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

type Screego struct {
	Image docker.Image
}

func (h *Screego) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Image,
	)
}

func (h *Screego) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "screego/server",
					Tag:        h.Image.Tag,
				},
				TargetImage: h.Image,
			},
		),
	}, nil
}

func (h *Screego) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: h.Image,
	}, nil
}
