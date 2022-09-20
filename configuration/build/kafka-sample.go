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

type KafkaSample struct {
	Image docker.Image
}

func (k *KafkaSample) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaSample) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "http_server",
				GitRepo:         "https://github.com/bborbe/sample_kafka.git",
				SourceDirectory: "github.com/bborbe/sample_kafka",
				Package:         "github.com/bborbe/sample_kafka/http_server",
				Image:           k.Image,
			},
		),
	}, nil
}

func (k *KafkaSample) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
