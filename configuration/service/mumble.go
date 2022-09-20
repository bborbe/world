// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Mumble struct {
	SSH          *ssh.SSH
	Version      docker.Tag
	Requirements []world.Configuration
}

func (p *Mumble) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.SSH,
		p.Version,
	)
}

func (p *Mumble) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, p.Requirements...)
	result = append(result, p.mumble()...)
	return result, nil
}

func (p *Mumble) mumble() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/mumble",
		Tag:        p.Version,
	}
	port := network.PortStatic(64738)
	return world.Configurations{
		&DisablePostfix{
			SSH: p.SSH,
		},
		&build.Mumble{
			Image: image,
		},
		&Docker{
			SSH:  p.SSH,
			Name: "mumble",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:   "mumble",
					Memory: 1000,
					Image:  image,
					Ports: []Port{
						{
							HostPort:   port,
							DockerPort: port,
						},
					},
					Requires: []remote.ServiceName{
						"docker.service",
					},
					After: []remote.ServiceName{
						"docker.service",
					},
					TimeoutStartSec: "20s",
					TimeoutStopSec:  "20s",
				}, nil
			},
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      p.SSH,
			Port:     port,
			Protocol: network.TCP,
		}),
	}
}

func (p *Mumble) Applier() (world.Applier, error) {
	return nil, nil
}
