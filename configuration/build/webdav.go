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

type Webdav struct {
	Image     docker.Image
	GitBranch docker.GitBranch
}

func (w *Webdav) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}
func (w *Webdav) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/webdav.git",
				Image:     w.Image,
				GitBranch: w.GitBranch,
			},
		),
	}, nil
}

func (w *Webdav) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: w.Image,
	}, nil
}
