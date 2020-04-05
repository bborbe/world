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

type NfsProvisioner struct {
	Image docker.Image
}

func (n *NfsProvisioner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Image,
	)
}

func (n *NfsProvisioner) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/kubernetes_incubator/nfs-provisioner",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		),
	}
}

func (n *NfsProvisioner) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
