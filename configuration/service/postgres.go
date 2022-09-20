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

type Postgres struct {
	SSH          *ssh.SSH
	Requirements []world.Configuration

	Name                 remote.ServiceName
	PostgresVersion      docker.Tag
	PostgresDatabaseName string
	PostgresInitDbArgs   string
	PostgresUsername     deployer.SecretValue
	PostgresPassword     deployer.SecretValue
	Port                 network.Port
	DataPath             file.Path
}

func (p *Postgres) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.SSH,
		p.DataPath,
		p.Port,
	)
}

func (p *Postgres) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, p.Requirements...)
	result = append(result, p.postgres()...)
	return result, nil
}

func (p *Postgres) postgres() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/postgres",
		Tag:        p.PostgresVersion,
	}
	envFile := fmt.Sprintf("%s.environment", p.DataPath.String())
	return world.Configurations{
		&build.Postgres{
			Image: image,
		},
		&remote.File{
			SSH:  p.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				postgresUsername, err := p.PostgresUsername.Value(ctx)
				if err != nil {
					return nil, err
				}
				postgresPassword, err := p.PostgresPassword.Value(ctx)
				if err != nil {
					return nil, err
				}
				return EnvFile{
					"POSTGRES_INITDB_ARGS": p.PostgresInitDbArgs,
					"PGDATA":               "/var/lib/postgresql/data/pgdata",
					"POSTGRES_DB":          p.PostgresDatabaseName,
					"POSTGRES_USER":        string(postgresUsername),
					"POSTGRES_PASSWORD":    string(postgresPassword),
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  p.SSH,
			Name: p.Name,
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				port, err := p.Port.Port(ctx)
				if err != nil {
					return nil, err
				}
				path, err := p.DataPath.Path(ctx)
				if err != nil {
					return nil, err
				}
				return &DockerServiceContent{
					Name:    p.Name,
					Memory:  200,
					Image:   image,
					Args:    []string{"postgres", "-c", "max_connections=150", "-p", strconv.Itoa(port)},
					HostNet: true,
					Volumes: []Volume{
						{
							HostPath:   path,
							DockerPath: "/var/lib/postgresql/data",
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
	}
}

func (p *Postgres) Applier() (world.Applier, error) {
	return nil, nil
}
