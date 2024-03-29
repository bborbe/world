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

type Minecraft struct {
	Image docker.Image
}

func (m *Minecraft) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Image,
	)
}

func (m *Minecraft) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/itzg/docker-minecraft-server",
				GitBranch: "master",
				Image:     m.Image,
			},
		),
	}, nil
}

func (m *Minecraft) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
