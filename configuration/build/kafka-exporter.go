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

type KafkaExporter struct {
	Image docker.Image
}

func (k *KafkaExporter) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Image,
	)
}

func (k *KafkaExporter) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "kafka-exporter",
				GitRepo:         "https://github.com/danielqsj/kafka_exporter.git",
				SourceDirectory: "github.com/danielqsj/kafka_exporter",
				Package:         "github.com/danielqsj/kafka_exporter",
				Image:           k.Image,
			},
		),
	}, nil
}

func (k *KafkaExporter) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: k.Image,
	}, nil
}
