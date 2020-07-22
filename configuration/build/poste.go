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

type Poste struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (p *Poste) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Image,
		p.VendorVersion,
		p.GitBranch,
	)
}

func (p *Poste) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/poste.io.git",
				Image:   p.Image,
				BuildArgs: docker.BuildArgs{
					"VENDOR_VERSION": p.VendorVersion.String(),
				},
				GitBranch: p.GitBranch,
			},
		),
	}
}

func (p *Poste) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
