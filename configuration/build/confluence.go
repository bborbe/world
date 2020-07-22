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

type Confluence struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (c *Confluence) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Image,
		c.VendorVersion,
		c.GitBranch,
	)
}

func (c *Confluence) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/atlassian-confluence.git",
				Image:   c.Image,
				BuildArgs: docker.BuildArgs{
					"VENDOR_VERSION": c.VendorVersion.String(),
				},
				GitBranch: c.GitBranch,
			},
		),
	}
}

func (c *Confluence) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
