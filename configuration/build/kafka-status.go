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

type KafkaStatus struct {
	Image docker.Image
}

func (k *KafkaStatus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaStatus) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "kafka-status",
				GitRepo:         "https://github.com/bborbe/kafka-status.git",
				SourceDirectory: "github.com/bborbe/kafka-status",
				Package:         "github.com/bborbe/kafka-status",
				Image:           k.Image,
			},
		),
	}
}

func (k *KafkaStatus) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
