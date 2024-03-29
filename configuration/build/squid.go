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

type Squid struct {
	Image docker.Image
}

func (s *Squid) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Image,
	)
}

func (s *Squid) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/squid.git",
				Image:     s.Image,
				GitBranch: docker.GitBranch(s.Image.Tag),
			},
		),
	}, nil
}

func (s *Squid) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: s.Image,
	}, nil
}
