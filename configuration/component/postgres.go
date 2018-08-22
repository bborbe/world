package component

import (
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Postgres struct {
	Context              k8s.Context
	Namespace            k8s.NamespaceName
	DataNfsPath          k8s.PodNfsPath
	DataNfsServer        k8s.PodNfsServer
	BackupNfsPath        k8s.PodNfsPath
	BackupNfsServer      k8s.PodNfsServer
	PostgresVersion      docker.Tag
	PostgresDatabaseName string
	PostgresInitDbArgs   string
	PostgresUsername     deployer.SecretValue
	PostgresPassword     deployer.SecretValue
}

func (p *Postgres) Children() []world.Configuration {
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
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "2000m",
							Memory: "200Mi",
						},
						Requests: k8s.Resources{
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
					Mounts: []k8s.VolumeMount{
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
					Nfs: k8s.PodVolumeNfs{
						Path:   k8s.PodNfsPath(p.DataNfsPath),
						Server: k8s.PodNfsServer(p.DataNfsServer),
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
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "backup",
					Image: postgresBackupImage,
					Requirement: &build.PostgresBackup{
						Image: postgresBackupImage,
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "500m",
							Memory: "100Mi",
						},
						Requests: k8s.Resources{
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
					Mounts: []k8s.VolumeMount{
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
				{
					Name:  "cleanup",
					Image: backupCleanUpImage,
					Requirement: &build.BackupCleanupCron{
						Image: backupCleanUpImage,
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.Resources{
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
					Mounts: []k8s.VolumeMount{
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
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/confluence-postgres-backup",
						Server: k8s.PodNfsServer(p.DataNfsServer),
					},
				},
			},
		},
	}
}

func (p *Postgres) Applier() (world.Applier, error) {
	return nil, nil
}
