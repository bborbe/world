// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"strconv"
	"time"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Webdav struct {
	SSH            *ssh.SSH
	Port           network.Port
	WebdavPassword deployer.SecretValue
	Requirements   []world.Configuration
}

func (l *Webdav) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		l.SSH,
		l.Port,
		l.WebdavPassword,
	)
}

func (l *Webdav) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, l.Requirements...)
	result = append(result, l.webdav()...)
	return result
}

func (l *Webdav) webdav() []world.Configuration {
	version := "1.1.0"
	image := docker.Image{
		Repository: "bborbe/webdav",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	envFile := "/home/webdav.environment"
	return []world.Configuration{
		&build.Webdav{
			GitBranch: docker.GitBranch(version),
			Image:     image,
		},
		&remote.File{
			SSH:  l.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				webdavPassword, err := l.WebdavPassword.Value(ctx)
				if err != nil {
					return nil, err
				}
				port, err := l.Port.Port(ctx)
				if err != nil {
					return nil, err
				}
				return EnvFile{
					"WEBDAV_USERNAME": "bborbe",
					"WEBDAV_PASSWORD": string(webdavPassword),
					"PORT":            strconv.Itoa(port),
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  l.SSH,
			Name: "webdav",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:    "webdav",
					Memory:  75,
					Image:   image,
					HostNet: true,
					Volumes: []Volume{
						{
							HostPath:   "/home/webdav",
							DockerPath: "/data/webdav",
						},
					},
					EnvironmentFiles: []string{
						envFile,
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
			SSH:  l.SSH,
			Port: l.Port,
		}),
	}
}

func (l *Webdav) Applier() (world.Applier, error) {
	return nil, nil
}
