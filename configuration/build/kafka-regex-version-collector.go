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

type KafkaRegexVersionCollector struct {
	Image docker.Image
}

func (k *KafkaRegexVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaRegexVersionCollector) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "kafka-regex-version-collector",
				GitRepo:         "https://github.com/bborbe/kafka-regex-version-collector.git",
				SourceDirectory: "github.com/bborbe/kafka-regex-version-collector",
				Package:         "github.com/bborbe/kafka-regex-version-collector",
				Image:           k.Image,
			},
		),
	}
}

func (k *KafkaRegexVersionCollector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
