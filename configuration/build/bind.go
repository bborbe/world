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

type Bind struct {
	Image     docker.Image
	GitBranch docker.GitBranch
}

func (b *Bind) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *Bind) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/bind.git",
				GitBranch: b.GitBranch,
				Image:     b.Image,
			},
		),
	}
}

func (b *Bind) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
