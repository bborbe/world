// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"github.com/bborbe/world/pkg/network"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ip struct {
	SSH  *ssh.SSH
	Tag  docker.Tag
	Port network.Port
}

func (i *Ip) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.SSH,
		i.Tag,
		i.Port,
	)
}

func (i *Ip) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/ip",
		Tag:        i.Tag,
	}
	return []world.Configuration{
		&build.Ip{
			Image: image,
		},
		&Docker{
			SSH:  i.SSH,
			Name: "etcd",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:   "ip",
					Memory: 100,
					Image:  image,
					Ports: []Port{
						{
							HostPort:   i.Port,
							DockerPort: network.PortStatic(8080),
						},
					},
					Args: []string{
						"-logtostderr",
						"-v=2",
						"--port=8080",
					},
					Requires: []remote.ServiceName{
						"docker.service",
					},
					After: []remote.ServiceName{
						"docker.service",
					},
					Before:          nil,
					TimeoutStartSec: "20s",
					TimeoutStopSec:  "20s",
				}, nil
			},
		},
	}
}

func (i *Ip) Applier() (world.Applier, error) {
	return nil, nil
}
