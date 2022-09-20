// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"time"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Bind struct {
	SSH *ssh.SSH
	IP  network.IP
}

func (e *Bind) Children(ctx context.Context) (world.Configurations, error) {
	version := "1.2.0"
	image := docker.Image{
		Repository: "bborbe/bind",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	return world.Configurations{
		&build.Bind{
			Image:     image,
			GitBranch: docker.GitBranch(version),
		},
		&Docker{
			SSH:  e.SSH,
			Name: "bind",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:    "bind",
					Memory:  256,
					HostNet: true,
					Volumes: []Volume{
						{
							HostPath:   "/data/bind",
							DockerPath: "/etc/bind",
						},
						{
							HostPath:   "/data/bind",
							DockerPath: "/var/lib/bind",
						},
					},
					Image: image,
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
	}, nil
}

func (e *Bind) Applier() (world.Applier, error) {
	return nil, nil
}

func (e *Bind) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.SSH,
		e.IP,
	)
}
