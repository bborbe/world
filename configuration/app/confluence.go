// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Confluence struct {
	Context          k8s.Context
	Domains          k8s.IngressHosts
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
	Requirements     []world.Configuration
}

func (c *Confluence) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.Domains,
		c.Version,
		c.DatabasePassword,
		c.SmtpPassword,
		c.SmtpUsername,
	)
}

func (c *Confluence) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, c.Requirements...)
	result = append(result, c.confluence()...)
	return result
}

func (c *Confluence) confluence() []world.Configuration {
	var buildVersion docker.GitBranch = "1.4.2"
	image := docker.Image{
		Repository: "bborbe/atlassian-confluence",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", c.Version, buildVersion)),
	}
	port := deployer.Port{
		Port:     8080,
		Protocol: "TCP",
		Name:     "http",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: c.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "confluence",
					Name:      "confluence",
				},
			},
		},
		&component.Postgres{
			Context:              c.Context,
			Namespace:            "confluence",
			DataPath:             "/data/confluence-postgres",
			BackupPath:           "/data/confluence-postgres-backup",
			PostgresVersion:      "10.13",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
			PostgresDatabaseName: "confluence",
			PostgresUsername:     deployer.SecretValueStatic("confluence"),
			PostgresPassword:     c.DatabasePassword,
		},
		&deployer.DeploymentDeployer{
			Context:      c.Context,
			Namespace:    "confluence",
			Name:         "confluence",
			Requirements: c.smtp().Requirements(),
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "confluence",
					Image: image,
					Requirement: &build.Confluence{
						VendorVersion: c.Version,
						GitBranch:     buildVersion,
						Image:         image,
					},
					Ports: []deployer.Port{port},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "4000m",
							Memory: "3000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "2000Mi",
						},
					},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "443",
						},
						{
							Name:  "SCHEMA",
							Value: "https",
						},
						{
							Name:  "HOSTNAME",
							Value: c.Domains[0].String(),
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "data",
							Path: "/var/lib/confluence",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 300,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						TimeoutSeconds:      5,
					},
				},
				c.smtp().Container(),
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Host: k8s.PodVolumeHost{
						Path: "/data/confluence-data",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   c.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			c.Context,
			"confluence",
			"confluence",
			"confluence",
			"http",
			"/",
			c.Domains...,
		),
	}
}

func (c *Confluence) smtp() *container.Smtp {
	return &container.Smtp{
		Hostname:     container.SmtpHostname(c.Domains[0].String()),
		Context:      c.Context,
		Namespace:    "confluence",
		SmtpPassword: c.SmtpPassword,
		SmtpUsername: c.SmtpUsername,
	}
}

func (c *Confluence) Applier() (world.Applier, error) {
	return nil, nil
}
