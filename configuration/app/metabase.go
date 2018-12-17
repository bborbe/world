// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Metabase struct {
	Context          k8s.Context
	NfsServer        k8s.PodNfsServer
	Domain           k8s.IngressHost
	DatabasePassword deployer.SecretValue
	Requirements     []world.Configuration
	NfsPrefix        string
}

func (t *Metabase) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Context,
		t.NfsServer,
		t.Domain,
		t.DatabasePassword,
	)
}

func (g *Metabase) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, g.Requirements...)
	result = append(result, g.metabase()...)
	return result
}

func (t *Metabase) metabase() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/metabase",
		Tag:        "v0.31.1",
	}
	port := deployer.Port{
		Port:     3000,
		Protocol: "TCP",
		Name:     "http",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: t.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "metabase",
					Name:      "metabase",
				},
			},
		},
		&component.Postgres{
			Context:              t.Context,
			Namespace:            "metabase",
			DataNfsPath:          k8s.PodNfsPath(t.NfsPrefix + "/metabase-postgres"),
			DataNfsServer:        t.NfsServer,
			BackupNfsPath:        k8s.PodNfsPath(t.NfsPrefix + "/metabase-postgres-backup"),
			BackupNfsServer:      t.NfsServer,
			PostgresVersion:      "10.5",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
			PostgresDatabaseName: "metabase",
			PostgresUsername: &deployer.SecretValueStatic{
				Content: []byte("metabase"),
			},
			PostgresPassword: t.DatabasePassword,
		},
		&deployer.SecretDeployer{
			Context:   t.Context,
			Namespace: "metabase",
			Name:      "metabase",
			Secrets: deployer.Secrets{
				"database-password": t.DatabasePassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   t.Context,
			Namespace: "metabase",
			Name:      "metabase",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "metabase",
					Image: image,
					Requirement: &build.Metabase{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1000m",
							Memory: "1500Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "300Mi",
						},
					},
					Env: []k8s.Env{
						{
							Name:  "JAVA_TOOL_OPTIONS",
							Value: "-Xmx1g",
						},
						{
							Name:  "MB_DB_TYPE",
							Value: "postgres",
						},
						{
							Name:  "MB_DB_DBNAME",
							Value: "metabase",
						},
						{
							Name:  "MB_DB_PORT",
							Value: "5432",
						},
						{
							Name:  "MB_DB_USER",
							Value: "metabase",
						},
						{
							Name:  "MB_DB_HOST",
							Value: "postgres",
						},
						{
							Name: "MB_DB_PASS",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "database-password",
									Name: "metabase",
								},
							},
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						FailureThreshold:    3,
						InitialDelaySeconds: 30,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						FailureThreshold:    1,
						InitialDelaySeconds: 10,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      5,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   t.Context,
			Namespace: "metabase",
			Name:      "metabase",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   t.Context,
			Namespace: "metabase",
			Name:      "metabase",
			Port:      "http",
			Domains:   k8s.IngressHosts{t.Domain},
		},
	}
}

func (t *Metabase) Applier() (world.Applier, error) {
	return nil, nil
}
