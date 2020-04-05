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

type Jira struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (j *Jira) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		j.Image,
		j.VendorVersion,
		j.GitBranch,
	)
}

func (j *Jira) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/atlassian-jira-software.git",
				Image:   j.Image,
				BuildArgs: docker.BuildArgs{
					"VENDOR_VERSION": j.VendorVersion.String(),
				},
				GitBranch: j.GitBranch,
			},
		),
	}
}

func (j *Jira) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: j.Image,
	}, nil
}
