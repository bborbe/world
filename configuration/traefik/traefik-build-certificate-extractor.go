// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traefik

import (
	"context"

	"github.com/bborbe/world/pkg/build"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BuildCertificateExtractor struct {
	Image docker.Image
}

func (b *BuildCertificateExtractor) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *BuildCertificateExtractor) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/DanielHuisman/traefik-certificate-extractor.git",
				GitBranch: docker.GitBranch(b.Image.Tag),
				Image:     b.Image,
			},
		),
	}
}

func (b *BuildCertificateExtractor) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
