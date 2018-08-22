package app

import (
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Jira struct {
	Cluster          cluster.Cluster
	Domains          []k8s.IngressHost
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
}

func (j *Jira) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.2.0"
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/atlassian-jira-software",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", j.Version, buildVersion)),
	}
	ports := []deployer.Port{
		{
			Port:     8080,
			Protocol: "TCP",
			Name:     "http",
		},
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
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "jira",
					Image: image,
					Requirement: &build.Jira{
						VendorVersion: j.Version,
						GitBranch:     buildVersion,
						Image:         image,
					},
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "4000m",
							Memory: "3000Mi",
						},
						Requests: k8s.Resources{
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
					Mounts: []k8s.VolumeMount{
						{
							Name: "data",
							Path: "/var/lib/jira",
						},
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
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jira",
			Name:      "jira",
			Port:      "http",
			Domains:   j.Domains,
		},
	}
}

func (j *Jira) smtp() *container.SmtpProvider {
	return &container.SmtpProvider{
		Hostname:     j.Domains[0].String(),
		Context:      j.Cluster.Context,
		Namespace:    "jira",
		SmtpPassword: j.SmtpPassword,
		SmtpUsername: j.SmtpUsername,
	}
}

func (j *Jira) Applier() (world.Applier, error) {
	return nil, nil
}
