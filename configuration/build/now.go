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

type Now struct {
	Image docker.Image
}

func (n *Now) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Image,
	)
}
func (n *Now) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/now.git",
				Image:     n.Image,
				GitBranch: docker.GitBranch(n.Image.Tag),
			},
		),
	}
}

func (n *Now) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
