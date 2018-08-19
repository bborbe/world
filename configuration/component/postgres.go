package component

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type Postges struct {
	Context              k8s.Context
	Namespace            k8s.NamespaceName
	DataNfsPath          deployer.MountNfsPath
	DataNfsServer        deployer.MountNfsServer
	BackupNfsPath        deployer.MountNfsPath
	BackupNfsServer      deployer.MountNfsServer
	PostgresVersion      docker.Tag
	PostgresDatabaseName string
	PostgresInitDbArgs   string
	PostgresUsername     deployer.SecretValue
	PostgresPassword     deployer.SecretValue
}

func (p *Postges) Children() []world.Configuration {
	postgresImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/postgres",
		Tag:        p.PostgresVersion,
	}
	postgresBackupImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/postgres-backup",
		Tag:        "2.0.1",
	}
	backupCleanUpImage := docker.Image{
		Registry:   "docker.io",
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
		&deployer.SecretDeployer{
			Context:   p.Context,
			Namespace: p.Namespace,
			Name:      "postgres",
			Secrets: deployer.Secrets{
				"username": p.PostgresUsername,
				"password": p.PostgresPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: p.Namespace,
			Name:      "postgres",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "postgres",
					Image: postgresImage,
					Requirement: &build.Postgres{
						Image: postgresImage,
					},
					Ports:         ports,
					CpuLimit:      "2000m",
					MemoryLimit:   "200Mi",
					CpuRequest:    "10m",
					MemoryRequest: "100Mi",
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
							Value: p.PostgresDatabaseName,
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
					Mounts: []deployer.Mount{
						{
							Name:   "data",
							Target: "/var/lib/postgresql/data",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "data",
					NfsServer: p.DataNfsServer,
					NfsPath:   p.DataNfsPath,
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
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "backup",
					Image: postgresBackupImage,
					Requirement: &build.PostgresBackup{
						Image: postgresBackupImage,
					},
					CpuLimit:      "500m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=1"},
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
							Value: p.PostgresDatabaseName,
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
					Mounts: []deployer.Mount{
						{
							Name:   "backup",
							Target: "/backup",
						},
					},
				},
				{
					Name:  "cleanup",
					Image: backupCleanUpImage,
					Requirement: &build.BackupCleanupCron{
						Image: backupCleanUpImage,
					},
					CpuLimit:      "100m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=1"},
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
					Mounts: []deployer.Mount{
						{
							Name:   "backup",
							Target: "/backup",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "backup",
					NfsPath:   "/data/confluence-postgres-backup",
					NfsServer: p.DataNfsServer,
				},
			},
		},
	}
}

func (p *Postges) Applier() world.Applier {
	return nil
}

func (p *Postges) Validate(ctx context.Context) error {
	if p.Context == "" {
		return errors.New("Context empty")
	}
	if p.Namespace == "" {
		return errors.New("Namespace empty")
	}
	if p.DataNfsPath == "" {
		return errors.New("DataNfsPath empty")
	}
	if p.DataNfsServer == "" {
		return errors.New("DataNfsServer empty")
	}
	if p.BackupNfsPath == "" {
		return errors.New("BackupNfsPath empty")
	}
	if p.BackupNfsServer == "" {
		return errors.New("BackupNfsServer empty")
	}
	if p.PostgresVersion == "" {
		return errors.New("PostgresVersion empty")
	}
	if p.PostgresDatabaseName == "" {
		return errors.New("PostgresDatabaseName empty")
	}
	if err := p.PostgresUsername.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	if err := p.PostgresPassword.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	return nil
}
