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

type KafkaLatestVersions struct {
	Image docker.Image
}

func (k *KafkaLatestVersions) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaLatestVersions) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "kafka-latest-versions",
				GitRepo:         "https://github.com/bborbe/kafka-latest-versions.git",
				SourceDirectory: "github.com/bborbe/kafka-latest-versions",
				Package:         "github.com/bborbe/kafka-latest-versions",
				Image:           k.Image,
			},
		),
	}
}

func (k *KafkaLatestVersions) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
