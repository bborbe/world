// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"
	"strconv"

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

type Confluence struct {
	SSH              *ssh.SSH
	AppPort          network.Port
	DBPort           network.Port
	Domain           network.Host
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
	Requirements     []world.Configuration
}

func (c *Confluence) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.SSH,
		c.Domain,
		c.Version,
		c.DatabasePassword,
		c.SmtpPassword,
		c.SmtpUsername,
	)
}

func (c *Confluence) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, c.Requirements...)
	result = append(result, &Postgres{
		SSH:  c.SSH,
		Name: "confluence-postgres",
		// https://confluence.atlassian.com/conf715/supported-platforms-1096098750.html
		PostgresVersion:      "14.3",
		PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
		PostgresDatabaseName: "confluence",
		PostgresUsername:     deployer.SecretValueStatic("confluence"),
		PostgresPassword:     c.DatabasePassword,
		Port:                 c.DBPort,
		DataPath:             "/home/confluence-postgres",
	})
	result = append(result, c.confluence()...)
	return result, nil
}

func (c *Confluence) confluence() []world.Configuration {
	var buildVersion docker.GitBranch = "1.5.3"
	image := docker.Image{
		Repository: "bborbe/atlassian-confluence",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", c.Version, buildVersion)),
	}
	path := "/home/confluence-data"
	envFile := fmt.Sprintf("%s.environment", path)
	memory := Memory(2048)
	return world.Configurations{
		&build.Confluence{
			VendorVersion: c.Version,
			GitBranch:     buildVersion,
			Image:         image,
		},
		&remote.File{
			SSH:  c.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				port, err := c.AppPort.Port(ctx)
				if err != nil {
					return nil, err
				}

				return EnvFile{
					"PROXY_NAME": c.Domain.String(),
					"PROXY_PORT": "443",
					"PORT":       strconv.Itoa(port),
					"ADDRESS":    "0.0.0.0",
					"SCHEMA":     "https",
					"MEMORY":     memory.String(),
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  c.SSH,
			Name: "confluence",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:    "confluence",
					Memory:  memory * 2,
					Image:   image,
					HostNet: true,
					Volumes: []Volume{
						{
							HostPath:   path,
							DockerPath: "/var/lib/confluence",
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
			SSH:      c.SSH,
			Port:     c.AppPort,
			Protocol: network.TCP,
		}),
	}
}

func (c *Confluence) Applier() (world.Applier, error) {
	return nil, nil
}
