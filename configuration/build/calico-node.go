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

type CalicoNode struct {
	Image docker.Image
}

func (t *CalicoNode) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *CalicoNode) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/calico/node",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *CalicoNode) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}