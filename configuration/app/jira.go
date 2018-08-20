package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Jira struct {
	Cluster          cluster.Cluster
	Domains          []deployer.Domain
	Version          docker.Tag
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
}

func (c *Jira) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.2.0"
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/atlassian-jira-software",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", c.Version, buildVersion)),
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
			Context:   c.Cluster.Context,
			Namespace: "jira",
		},
		&component.Postgres{
			Context:              c.Cluster.Context,
			Namespace:            "jira",
			DataNfsPath:          "/data/jira-postgres",
			DataNfsServer:        c.Cluster.NfsServer,
			BackupNfsPath:        "/data/jira-postgres-backup",
			BackupNfsServer:      c.Cluster.NfsServer,
			PostgresVersion:      "9.6-alpine",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=POSIX.UTF-8 --lc-ctype=POSIX.UTF-8 -T",
			PostgresDatabaseName: "jira",
			PostgresUsername: &deployer.SecretValueStatic{
				Content: []byte("jira"),
			},
			PostgresPassword: c.DatabasePassword,
		},
		&deployer.DeploymentDeployer{
			Context:      c.Cluster.Context,
			Namespace:    "jira",
			Name:         "jira",
			Requirements: c.smtp().Requirements(),
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "jira",
					Image: image,
					Requirement: &build.Jira{
						VendorVersion: c.Version,
						GitBranch:     buildVersion,
						Image:         image,
					},
					Ports:         ports,
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
							Target: "/var/lib/jira",
						},
					},
				},
				c.smtp().Container(),
			},
			Volumes: []deployer.Volume{
				{
					Name:      "data",
					NfsPath:   "/data/jira-data",
					NfsServer: c.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "jira",
			Name:      "jira",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   c.Cluster.Context,
			Namespace: "jira",
			Name:      "jira",
			Domains:   c.Domains,
		},
	}
}

func (c *Jira) smtp() *container.SmtpProvider {
	return &container.SmtpProvider{
		Hostname:     c.Domains[0].String(),
		Context:      c.Cluster.Context,
		Namespace:    "jira",
		SmtpPassword: c.SmtpPassword,
		SmtpUsername: c.SmtpUsername,
	}
}

func (c *Jira) Applier() world.Applier {
	return nil
}

func (c *Jira) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate jira app ...")
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
	if err := c.smtp().Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	return nil
}
