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

type HelloWorld struct {
	Image docker.Image
}

func (t *HelloWorld) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (h *HelloWorld) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/hello-world.git",
				Image:     h.Image,
				GitBranch: docker.GitBranch(h.Image.Tag),
			},
		},
	}
}

func (h *HelloWorld) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: h.Image,
	}, nil
}
