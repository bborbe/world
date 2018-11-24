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

type DebugServer struct {
	Image docker.Image
}

func (t *DebugServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (p *DebugServer) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "debug-server",
				GitRepo:         "https://github.com/bborbe/debug-server.git",
				SourceDirectory: "github.com/bborbe/debug-server",
				Package:         "github.com/bborbe/debug-server",
				Image:           p.Image,
			},
		},
	}
}

func (p *DebugServer) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
