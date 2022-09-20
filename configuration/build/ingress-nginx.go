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

type Build struct {
	Image docker.Image
}

func (b *Build) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *Build) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/kubernetes-ingress-controller/nginx-ingress-controller",
					Tag:        b.Image.Tag,
				},
				TargetImage: b.Image,
			},
		),
	}, nil
}

func (b *Build) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
