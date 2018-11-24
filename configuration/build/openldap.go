// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Openldap struct {
	Image docker.Image
}

func (o *Openldap) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/openldap.git",
				Image:     o.Image,
				GitBranch: docker.GitBranch(o.Image.Tag),
			},
		},
	}
}

func (o *Openldap) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: o.Image,
	}, nil
}

func (d *Openldap) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Image,
	)
}
