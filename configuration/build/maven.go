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

type Maven struct {
	Image docker.Image
}

func (m *Maven) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Image,
	)
}

func (m *Maven) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "maven",
				GitRepo:         "https://github.com/bborbe/maven.git",
				SourceDirectory: "github.com/bborbe/maven",
				Package:         "github.com/bborbe/maven",
				Image:           m.Image,
			},
		),
	}, nil
}

func (m *Maven) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
