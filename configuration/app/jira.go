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

type Jira struct {
	Cluster          cluster.Cluster
	Domain           k8s.IngressHost
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
}

func (t *Jira) Validate(ctx context.Context) error {
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

func (j *Jira) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.2.0"
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
		&deployer.NamespaceDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jira",
		},
		&component.Postgres{
			Context:              j.Cluster.Context,
			Namespace:            "jira",
			DataNfsPath:          "/data/jira-postgres",
			DataNfsServer:        j.Cluster.NfsServer,
			BackupNfsPath:        "/data/jira-postgres-backup",
			BackupNfsServer:      j.Cluster.NfsServer,
			PostgresVersion:      "9.6-alpine",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=POSIX.UTF-8 --lc-ctype=POSIX.UTF-8 -T",
			PostgresDatabaseName: "jira",
			PostgresUsername: &deployer.SecretValueStatic{
				Content: []byte("jira"),
			},
			PostgresPassword: j.DatabasePassword,
		},
		&deployer.DeploymentDeployer{
			Context:      j.Cluster.Context,
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
							Value: j.Domain.String(),
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
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/jira-data",
						Server: j.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jira",
			Name:      "jira",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jira",
			Name:      "jira",
			Port:      "http",
			Domains:   k8s.IngressHosts{j.Domain},
		},
	}
}

func (j *Jira) smtp() *container.Smtp {
	return &container.Smtp{
		Hostname:     container.SmtpHostname(j.Domain.String()),
		Context:      j.Cluster.Context,
		Namespace:    "jira",
		SmtpPassword: j.SmtpPassword,
		SmtpUsername: j.SmtpUsername,
	}
}

func (j *Jira) Applier() (world.Applier, error) {
	return nil, nil
}
