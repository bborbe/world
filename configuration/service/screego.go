// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Screego struct {
	SSH     *ssh.SSH
	Version docker.Tag
	IP      network.IP
}

func (s *Screego) Children(ctx context.Context) (world.Configurations, error) {
	image := docker.Image{
		Repository: "bborbe/screego",
		Tag:        s.Version,
	}
	envFile := "/data/screego/environment"
	usersFile := "/data/screego/users"
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      s.SSH,
			Port:     network.PortStatic(3478),
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      s.SSH,
			Port:     network.PortStatic(5050),
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH: s.SSH,
			PortRange: &network.PortRange{
				From: network.PortStatic(50000),
				To:   network.PortStatic(50100),
			},
			Protocol: network.UDP,
		}),
		&build.Screego{
			Image: image,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Directory{
			SSH:  s.SSH,
			Path: file.Path("/data/screego"),
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   s.SSH,
			Path:  file.Path("/data/screego"),
			User:  "root",
			Group: "root",
		}),
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path(usersFile),
			// screego hash --name "user1" --pass "your password"
			Content: content.Static(`seibert:$2a$12$/bHGYtaWhkNZcQGci5jWq.3s3Rf2FybONkheQTN6cZgeQ3CnRKafK`),
			User:    "root",
			Group:   "root",
			Perm:    0644,
		},
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				ip, err := s.IP.IP(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "get ip failed")
				}
				return EnvFile{
					"SCREEGO_TURN_PORT_RANGE":              "50000:50100",
					"SCREEGO_TRUST_PROXY_HEADERS":          "true",
					"SCREEGO_TURN_STRICT_AUTH":             "false",
					"SCREEGO_EXTERNAL_IP":                  ip.String(),
					"SCREEGO_USERS_FILE":                   "/data/screego/users",
					"SCREEGO_CLOSE_ROOM_WHEN_OWNER_LEAVES": "false",
					"SCREEGO_SECRET":                       "ae63d9478b172ac8499a18344f9fb0c3",
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  s.SSH,
			Name: "screego",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:            "screego",
					Memory:          2048,
					Image:           image,
					TimeoutStartSec: "20s",
					TimeoutStopSec:  "20s",
					EnvironmentFiles: []string{
						envFile,
					},
					Volumes: []Volume{
						{
							HostPath:   "/data/screego",
							DockerPath: "/data/screego",
						},
					},
					HostNet: true,
				}, nil
			},
		},
	}, nil
}

func (s *Screego) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Screego) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Version,
		s.IP,
	)
}
