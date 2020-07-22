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

type NginxAutoindex struct {
	Image docker.Image
}

func (n *NginxAutoindex) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Image,
	)
}

func (n *NginxAutoindex) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "jrelva/nginx-autoindex",
					Tag:        "latest",
				},
				TargetImage: n.Image,
			},
		),
	}
}

func (n *NginxAutoindex) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
