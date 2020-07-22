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

type MqttKafkaConnector struct {
	Image docker.Image
}

func (m *MqttKafkaConnector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Image,
	)
}

func (m *MqttKafkaConnector) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "mqtt-kafka-connector",
				GitRepo:         "https://github.com/bborbe/mqtt-kafka-connector.git",
				SourceDirectory: "github.com/bborbe/mqtt-kafka-connector",
				Package:         "github.com/bborbe/mqtt-kafka-connector",
				Image:           m.Image,
			},
		),
	}
}

func (m *MqttKafkaConnector) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: m.Image,
	}, nil
}
