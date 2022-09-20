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

type Ip struct {
	Image docker.Image
}

func (i *Ip) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.Image,
	)
}

func (i *Ip) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "ip",
				GitRepo:         "https://github.com/bborbe/ip.git",
				SourceDirectory: "github.com/bborbe/ip",
				Package:         "github.com/bborbe/ip/cmd/ip-server",
				Image:           i.Image,
			},
		),
	}, nil
}

func (i *Ip) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
