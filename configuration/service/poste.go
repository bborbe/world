// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"

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

type Poste struct {
	SSH          *ssh.SSH
	PosteVersion docker.Tag
	Port         network.Port
	Requirements []world.Configuration
}

func (p *Poste) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.SSH,
		p.PosteVersion,
		p.Port,
	)
}

func (p *Poste) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, p.Requirements...)
	result = append(result, p.poste()...)
	return result, nil
}

func (p *Poste) poste() []world.Configuration {

	smtpPort := network.PortStatic(25)
	smtptlsPort := network.PortStatic(465)
	smtpsPort := network.PortStatic(587)
	imapsPort := network.PortStatic(993)

	var buildVersion docker.GitBranch = "2.0.1"
	image := docker.Image{
		Repository: "bborbe/poste.io",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", p.PosteVersion, buildVersion)),
	}
	envFile := "/home/poste.environment"
	return world.Configurations{
		&DisablePostfix{
			SSH: p.SSH,
		},
		&build.Poste{
			Image:         image,
			GitBranch:     buildVersion,
			VendorVersion: p.PosteVersion,
		},
		&remote.File{
			SSH:  p.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				return EnvFile{
					"HTTPS": "OFF",
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  p.SSH,
			Name: "poste",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:   "poste",
					Memory: 1000,
					Image:  image,
					Ports: []Port{
						{
							HostPort:   smtpPort,
							DockerPort: smtpPort,
						},
						{
							HostPort:   smtpsPort,
							DockerPort: smtpsPort,
						},
						{
							HostPort:   smtptlsPort,
							DockerPort: smtptlsPort,
						},
						{
							HostPort:   imapsPort,
							DockerPort: imapsPort,
						},
						{
							HostPort:   p.Port,
							DockerPort: network.PortStatic(80),
						},
					},
					Volumes: []Volume{
						{
							HostPath:   "/data/poste",
							DockerPath: "/data",
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
			SSH:      p.SSH,
			Port:     smtpPort,
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      p.SSH,
			Port:     smtpsPort,
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      p.SSH,
			Port:     smtptlsPort,
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      p.SSH,
			Port:     imapsPort,
			Protocol: network.TCP,
		}),
	}
}

func (p *Poste) Applier() (world.Applier, error) {
	return nil, nil
}
