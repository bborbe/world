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

type GitSync struct {
	Image docker.Image
}

func (g *GitSync) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		g.Image,
	)
}

func (g *GitSync) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/git-sync.git",
				Image:     g.Image,
				GitBranch: docker.GitBranch(g.Image.Tag),
			},
		),
	}
}

func (g *GitSync) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: g.Image,
	}, nil
}
