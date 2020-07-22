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

type Jira struct {
	Context          k8s.Context
	Domains          k8s.IngressHosts
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
	Requirements     []world.Configuration
}

func (j *Jira) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		j.Context,
		j.Domains,
		j.Version,
		j.DatabasePassword,
		j.SmtpPassword,
		j.SmtpUsername,
	)
}

func (j *Jira) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, j.Requirements...)
	result = append(result, j.jira()...)
	return result
}

func (j *Jira) jira() []world.Configuration {
	var buildVersion docker.GitBranch = "1.3.2"
	image := docker.Image{
		Repository: "bborbe/atlassian-jira-software",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", j.Version, buildVersion)),
	}
	port := deployer.Port{
		Port:     8080,
		Protocol: "TCP",
		Name:     "http",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: j.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "jira",
					Name:      "jira",
				},
			},
		},
		&component.Postgres{
			Context:              j.Context,
			Namespace:            "jira",
			DataPath:             "/data/jira-postgres",
			BackupPath:           "/data/jira-postgres-backup",
			PostgresVersion:      "10.13",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=C.UTF-8 --lc-ctype=C.UTF-8 -T template0",
			PostgresDatabaseName: "jira",
			PostgresUsername:     deployer.SecretValueStatic("jira"),
			PostgresPassword:     j.DatabasePassword,
		},
		&deployer.DeploymentDeployer{
			Context:      j.Context,
			Namespace:    "jira",
			Name:         "jira",
			Requirements: j.smtp().Requirements(),
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "jira",
					Image: image,
					Requirement: &build.Jira{
						VendorVersion: j.Version,
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
							Value: j.Domains[0].String(),
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "data",
							Path: "/var/lib/jira",
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
				j.smtp().Container(),
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Host: k8s.PodVolumeHost{
						Path: "/data/jira-data",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   j.Context,
			Namespace: "jira",
			Name:      "jira",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			j.Context,
			"jira",
			"jira",
			"jira",
			"http",
			"/",
			j.Domains...,
		),
	}
}

func (j *Jira) smtp() *container.Smtp {
	return &container.Smtp{
		Hostname:     container.SmtpHostname(j.Domains[0].String()),
		Context:      j.Context,
		Namespace:    "jira",
		SmtpPassword: j.SmtpPassword,
		SmtpUsername: j.SmtpUsername,
	}
}

func (j *Jira) Applier() (world.Applier, error) {
	return nil, nil
}
