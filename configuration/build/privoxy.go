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

type Privoxy struct {
	Image docker.Image
}

func (p *Privoxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Image,
	)
}

func (p *Privoxy) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/privoxy.git",
				Image:     p.Image,
				GitBranch: docker.GitBranch(p.Image.Tag),
			},
		),
	}, nil
}

func (p *Privoxy) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
