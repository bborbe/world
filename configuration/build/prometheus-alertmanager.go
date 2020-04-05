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

type PrometheusAlertmanager struct {
	Image docker.Image
}

func (p *PrometheusAlertmanager) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Image,
	)
}

func (p *PrometheusAlertmanager) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "quay.io/prometheus/alertmanager",
					Tag:        p.Image.Tag,
				},
				TargetImage: p.Image,
			},
		),
	}
}

func (p *PrometheusAlertmanager) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
