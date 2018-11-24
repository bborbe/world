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

type Teamvault struct {
	Image docker.Image
}

func (t *Teamvault) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (p *Teamvault) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/teamvault.git",
				Image:   p.Image,
				BuildArgs: docker.BuildArgs{
					"VERSION": p.Image.Tag.String(),
				},
				GitBranch: "master",
			},
		},
	}
}

func (p *Teamvault) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
