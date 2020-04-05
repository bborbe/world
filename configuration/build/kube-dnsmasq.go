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

type KubednsMasq struct {
	Image docker.Image
}

func (k *KubednsMasq) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KubednsMasq) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "gcr.io/google_containers/kube-dnsmasq-amd64",
					Tag:        k.Image.Tag,
				},
				TargetImage: k.Image,
			},
		),
	}
}

func (k *KubednsMasq) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
