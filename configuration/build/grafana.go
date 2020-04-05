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

type Grafana struct {
	Image docker.Image
}

func (g *Grafana) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		g.Image,
	)
}

func (g *Grafana) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "grafana/grafana",
					Tag:        g.Image.Tag,
				},
				TargetImage: g.Image,
			},
		),
	}
}

func (g *Grafana) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: g.Image,
	}, nil
}
