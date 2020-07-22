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

type Password struct {
	Image docker.Image
}

func (p *Password) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Image,
	)
}

func (p *Password) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "password",
				GitRepo:         "https://github.com/bborbe/password.git",
				SourceDirectory: "github.com/bborbe/password",
				Package:         "github.com/bborbe/password/cmd/password-server",
				Image:           p.Image,
			},
		),
	}
}

func (p *Password) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
