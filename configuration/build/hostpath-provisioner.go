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

type HostPathProvisioner struct {
	Image docker.Image
}

func (h *HostPathProvisioner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Image,
	)
}

func (h *HostPathProvisioner) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "hostpath-provisioner",
				GitRepo:         "https://github.com/bborbe/hostpath-provisioner.git",
				SourceDirectory: "github.com/bborbe/hostpath-provisioner",
				Package:         "github.com/bborbe/hostpath-provisioner",
				Image:           h.Image,
			},
		),
	}
}

func (h *HostPathProvisioner) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: h.Image,
	}, nil
}
