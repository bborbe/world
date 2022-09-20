// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ldap struct {
	SSH          *ssh.SSH
	Tag          docker.Tag
	LdapPassword deployer.SecretValue
	Requirements []world.Configuration
}

func (l *Ldap) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, l.Requirements...)
	result = append(result, l.ldap()...)
	return result, nil
}

func (l *Ldap) ldap() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/openldap",
		Tag:        l.Tag,
	}
	envFile := "/home/ldap.environment"
	return world.Configurations{
		&build.Openldap{
			Image: image,
		},
		&remote.File{
			SSH:  l.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				value, err := l.LdapPassword.Value(ctx)
				if err != nil {
					return nil, err
				}
				return EnvFile{
					"LDAP_SUFFIX": "dc=benjamin-borbe,dc=de",
					"LDAP_ROOTDN": "cn=root,dc=benjamin-borbe,dc=de",
					"LDAP_SECRET": string(value),
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  l.SSH,
			Name: "ldap",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:    "ldap",
					Memory:  75,
					Image:   image,
					HostNet: true,
					Volumes: []Volume{
						{
							HostPath:   "/home/ldap",
							DockerPath: "/var/lib/openldap/openldap-data",
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

func (l *Ldap) Applier() (world.Applier, error) {
	return nil, nil
}

func (l *Ldap) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		l.SSH,
	)
}
