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

type Erpnext struct {
	Image docker.Image
}

func (e *Erpnext) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.Image,
	)
}

func (e *Erpnext) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/erpnext.git",
				Image:     e.Image,
				GitBranch: docker.GitBranch(e.Image.Tag),
			},
		),
	}, nil
}

func (e *Erpnext) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: e.Image,
	}, nil
}
