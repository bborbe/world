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

type Mumble struct {
	Image docker.Image
}

func (m *Mumble) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Image,
	)
}

func (m *Mumble) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/mumble.git",
				Image:     m.Image,
				GitBranch: docker.GitBranch(m.Image.Tag),
			},
		),
	}
}

func (m *Mumble) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
