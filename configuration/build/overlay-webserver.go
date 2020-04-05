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

type OverlayWebserver struct {
	Image docker.Image
}

func (o *OverlayWebserver) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		o.Image,
	)
}

func (o *OverlayWebserver) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "overlay-server",
				GitRepo:         "https://github.com/bborbe/server.git",
				SourceDirectory: "github.com/bborbe/server",
				Package:         "github.com/bborbe/server/cmd/overlay-server",
				Image:           o.Image,
			},
		),
	}
}

func (o *OverlayWebserver) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}
