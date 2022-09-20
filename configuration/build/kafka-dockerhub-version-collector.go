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

type KafkaDockerhubVersionCollector struct {
	Image docker.Image
}

func (k *KafkaDockerhubVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaDockerhubVersionCollector) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "kafka-dockerhub-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-dockerhub-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-dockerhub-version-collector",
				Package:         "github.com/bborbe/kafka-dockerhub-version-collector",
				Image:           k.Image,
			},
		),
	}, nil
}

func (k *KafkaDockerhubVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
