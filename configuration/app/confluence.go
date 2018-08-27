package app

import (
	"fmt"

	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Confluence struct {
	Cluster          cluster.Cluster
	Domain           k8s.IngressHost
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
}

func (t *Confluence) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domain,
		t.Version,
		t.DatabasePassword,
		t.SmtpPassword,
		t.SmtpUsername,
	)
}

func (c *Confluence) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.3.0"
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
		&deployer.NamespaceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
		},
		&component.Postgres{
			Context:              c.Cluster.Context,
			Namespace:            "confluence",
			DataNfsPath:          "/data/confluence-postgres",
			DataNfsServer:        c.Cluster.NfsServer,
			BackupNfsPath:        "/data/confluence-postgres-backup",
			BackupNfsServer:      c.Cluster.NfsServer,
			PostgresVersion:      "9.5.14",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
			PostgresDatabaseName: "confluence",
			PostgresUsername: &deployer.SecretValueStatic{
				Content: []byte("confluence"),
			},
			PostgresPassword: c.DatabasePassword,
		},
		&deployer.DeploymentDeployer{
			Context:      c.Cluster.Context,
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
							Value: c.Domain.String(),
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
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/confluence-data",
						Server: k8s.PodNfsServer(c.Cluster.NfsServer),
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Port:      "http",
			Domains:   k8s.IngressHosts{c.Domain},
		},
	}
}

func (c *Confluence) smtp() *container.Smtp {
	return &container.Smtp{
		Hostname:     container.SmtpHostname(c.Domain.String()),
		Context:      c.Cluster.Context,
		Namespace:    "confluence",
		SmtpPassword: c.SmtpPassword,
		SmtpUsername: c.SmtpUsername,
	}
}

func (c *Confluence) Applier() (world.Applier, error) {
	return nil, nil
}
