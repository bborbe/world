package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Confluence struct {
	Cluster          cluster.Cluster
	Domains          []deployer.Domain
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
}

func (b *Confluence) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "confluence",
		},
	}
	result = append(result, b.postgres()...)
	result = append(result, b.app()...)
	return result
}

func (c *Confluence) app() []world.Configuration {
	var buildVersion docker.GitBranch = "1.3.0"
	confluenceImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/atlassian-confluence",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", c.Version, buildVersion)),
	}
	confluencePorts := []deployer.Port{
		{
			Port:     8080,
			Protocol: "TCP",
			Name:     "http",
		},
	}
	smtpImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/smtp",
		Tag:        "1.2.1",
	}
	smtpPorts := []deployer.Port{
		{
			Port:     25,
			Protocol: "TCP",
			Name:     "smtp",
		},
	}
	return []world.Configuration{
		&deployer.SecretDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Secrets: deployer.Secrets{
				"smtp-username": c.SmtpUsername,
				"smtp-password": c.SmtpPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Requirements: []world.Configuration{
				&build.Confluence{
					VendorVersion: c.Version,
					GitBranch:     buildVersion,
					Image:         confluenceImage,
				},
				&build.Smtp{
					Image: smtpImage,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "confluence",
					Image:         confluenceImage,
					Ports:         confluencePorts,
					CpuLimit:      "4000m",
					MemoryLimit:   "3000Mi",
					CpuRequest:    "100m",
					MemoryRequest: "2000Mi",
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
					Mounts: []deployer.Mount{
						{
							Name:   "data",
							Target: "/var/lib/confluence",
						},
					},
				},
				{
					Name:          "smtp",
					Image:         smtpImage,
					Ports:         smtpPorts,
					CpuLimit:      "250m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Env: []k8s.Env{
						{
							Name:  "HOSTNAME",
							Value: c.Domains[0].String(),
						},
						{
							Name:  "RELAY_SMTP_PORT",
							Value: "25",
						},
						{
							Name:  "RELAY_SMTP_SERVER",
							Value: "mail.benjamin-borbe.de",
						},
						{
							Name:  "RELAY_SMTP_TLS",
							Value: "false",
						},
						{
							Name:  "ALLOWED_SENDER_DOMAINS",
							Value: "",
						},
						{
							Name:  "ALLOWED_NETWORKS",
							Value: "",
						},
						{
							Name: "RELAY_SMTP_USERNAME",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "smtp-username",
									Name: "confluence",
								},
							},
						},
						{
							Name: "RELAY_SMTP_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "smtp-password",
									Name: "confluence",
								},
							},
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "data",
					NfsPath:   "/data/confluence-data",
					NfsServer: c.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Ports:     confluencePorts,
		},
		&deployer.IngressDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Domains:   c.Domains,
		},
	}
}

func (c *Confluence) postgres() []world.Configuration {
	postgresImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/postgres",
		Tag:        "9.5.14",
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
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "postgres",
			Secrets: deployer.Secrets{
				"password": c.DatabasePassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "postgres",
			Requirements: []world.Configuration{
				&build.Postgres{
					Image: postgresImage,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "postgres",
					Image:         postgresImage,
					Ports:         ports,
					CpuLimit:      "2000m",
					MemoryLimit:   "200Mi",
					CpuRequest:    "10m",
					MemoryRequest: "100Mi",
					Env: []k8s.Env{
						{
							Name: "POSTGRES_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "postgres",
								},
							},
						},
						{
							Name:  "PGDATA",
							Value: "/var/lib/postgresql/data/pgdata",
						},
						{
							Name:  "POSTGRES_USER",
							Value: "confluence",
						},
						{
							Name:  "POSTGRES_DB",
							Value: "confluence",
						},
						{
							Name:  "POSTGRES_INITDB_ARGS",
							Value: "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
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
					NfsPath:   "/data/confluence-postgres",
					NfsServer: c.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "postgres",
			Ports:     ports,
		},

		&deployer.DeploymentDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "postgres-backup",
			Requirements: []world.Configuration{
				&build.PostgresBackup{
					Image: postgresBackupImage,
				},
				&build.BackupCleanupCron{
					Image: backupCleanUpImage,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "backup",
					Image:         postgresBackupImage,
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
							Value: "confluence-postgres",
						},
						{
							Name:  "PORT",
							Value: "5432",
						},
						{
							Name:  "DATABASE",
							Value: "confluence",
						},
						{
							Name:  "USERNAME",
							Value: "confluence",
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
					Name:          "cleanup",
					Image:         backupCleanUpImage,
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
							Value: "postgres_confluence_.*.dump",
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
					NfsServer: c.Cluster.NfsServer,
				},
			},
		},
	}
}

func (c *Confluence) Applier() world.Applier {
	return nil
}

func (c *Confluence) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate confluence app ...")
	if err := c.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	if len(c.Domains) != 1 {
		return errors.New("need exact one domain")
	}
	if c.Version == "" {
		return errors.New("Tag empty")
	}
	if err := c.DatabasePassword.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	if err := c.SmtpUsername.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	if err := c.SmtpPassword.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	return nil
}
