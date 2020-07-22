// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package component

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type DatabaseName string

func (d DatabaseName) String() string {
	return string(d)
}

func (d DatabaseName) Validate(ctx context.Context) error {
	if d == "" {
		return errors.New("DatabaseName empty")
	}
	return nil
}

type Postgres struct {
	Context              k8s.Context
	Namespace            k8s.NamespaceName
	DataPath             k8s.PodHostPath
	BackupPath           k8s.PodHostPath
	PostgresVersion      docker.Tag
	PostgresDatabaseName DatabaseName
	PostgresInitDbArgs   string
	PostgresUsername     deployer.SecretValue
	PostgresPassword     deployer.SecretValue
}

func (t *Postgres) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Context,
		t.Namespace,
		t.DataPath,
		t.BackupPath,
		t.PostgresVersion,
		t.PostgresDatabaseName,
		t.PostgresUsername,
		t.PostgresPassword,
	)
}

func (p *Postgres) Children() []world.Configuration {
	postgresImage := docker.Image{
		Repository: "bborbe/postgres",
		Tag:        p.PostgresVersion,
	}
	postgresBackupImage := docker.Image{
		Repository: "bborbe/postgres-backup",
		Tag:        "2.0.1",
	}
	backupCleanUpImage := docker.Image{
		Repository: "bborbe/backup-cleanup",
		Tag:        "1.2.0",
	}
	ports := []deployer.Port{
		{
			Port:     5432,
			Protocol: "TCP",
			Name:     "postgres",
		},
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   p.Context,
				Namespace: p.Namespace,
				Name:      "postgres",
				Secrets: deployer.Secrets{
					"username": p.PostgresUsername,
					"password": p.PostgresPassword,
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: p.Namespace,
			Name:      "postgres",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "postgres",
					Image: postgresImage,
					Requirement: &build.Postgres{
						Image: postgresImage,
					},
					Ports: ports,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "2000m",
							Memory: "200Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "100Mi",
						},
					},
					Args: []k8s.Arg{"postgres", "-c", "max_connections=150"},
					Env: []k8s.Env{
						{
							Name:  "POSTGRES_INITDB_ARGS",
							Value: p.PostgresInitDbArgs,
						},
						{
							Name:  "PGDATA",
							Value: "/var/lib/postgresql/data/pgdata",
						},
						{
							Name:  "POSTGRES_DB",
							Value: p.PostgresDatabaseName.String(),
						},
						{
							Name: "POSTGRES_USER",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "username",
									Name: "postgres",
								},
							},
						},
						{
							Name: "POSTGRES_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "postgres",
								},
							},
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "data",
							Path: "/var/lib/postgresql/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Host: k8s.PodVolumeHost{
						Path: p.DataPath,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: p.Namespace,
			Name:      "postgres",
			Ports:     ports,
		},
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: p.Namespace,
			Name:      "postgres-backup",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "backup",
					Image: postgresBackupImage,
					Requirement: &build.PostgresBackup{
						Image: postgresBackupImage,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Env: []k8s.Env{
						{
							Name:  "LOCK",
							Value: "/postgres_backup_cron.lock",
						},
						{
							Name:  "HOST",
							Value: "postgres",
						},
						{
							Name:  "PORT",
							Value: "5432",
						},
						{
							Name:  "DATABASE",
							Value: p.PostgresDatabaseName.String(),
						},
						{
							Name: "USERNAME",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "username",
									Name: "postgres",
								},
							},
						},
						{
							Name: "PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "postgres",
								},
							},
						},
						{
							Name:  "TARGETDIR",
							Value: "/backup",
						},
						{
							Name:  "WAIT",
							Value: "1h",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
				&deployer.DeploymentDeployerContainer{
					Name:  "cleanup",
					Image: backupCleanUpImage,
					Requirement: &build.BackupCleanupCron{
						Image: backupCleanUpImage,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Env: []k8s.Env{
						{
							Name:  "LOCK",
							Value: "/backup_cleanup_cron.lock",
						},
						{
							Name:  "WAIT",
							Value: "1h",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "KEEP",
							Value: "5",
						},
						{
							Name:  "DIR",
							Value: "/backup",
						},
						{
							Name:  "MATCH",
							Value: fmt.Sprintf("postgres_%s_.*.dump", p.PostgresDatabaseName),
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Host: k8s.PodVolumeHost{
						Path: p.BackupPath,
					},
				},
			},
		},
	}
}

func (p *Postgres) Applier() (world.Applier, error) {
	return nil, nil
}
