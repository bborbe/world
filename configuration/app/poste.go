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

type Poste struct {
	Cluster      cluster.Cluster
	PosteVersion docker.Tag
	Domains      []deployer.Domain
}

func (p *Poste) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.0.0"
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/poste.io",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", p.PosteVersion, buildVersion)),
	}
	ports := []deployer.Port{
		{
			Name:     "smtp",
			Port:     25,
			HostPort: 25,
			Protocol: "TCP",
		},
		{
			Name:     "http",
			Port:     80,
			Protocol: "TCP",
		},
		{
			Name:     "pop3",
			Port:     110,
			Protocol: "TCP",
		},
		{
			Name:     "imap",
			Port:     143,
			Protocol: "TCP",
		},
		{
			Name:     "https",
			Port:     443,
			Protocol: "TCP",
		},
		{
			Name:     "smtptls",
			Port:     465,
			HostPort: 465,
			Protocol: "TCP",
		},
		{
			Name:     "smtps",
			Port:     587,
			HostPort: 587,
			Protocol: "TCP",
		},
		{
			Name:     "imaps",
			Port:     993,
			HostPort: 993,
			Protocol: "TCP",
		},
		{
			Name:     "pop3s",
			Port:     995,
			Protocol: "TCP",
		},
		{
			Name:     "sieve",
			Port:     4190,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "poste",
		},
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "poste",
			Name:      "poste",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "poste",
					CpuLimit:      "1500m",
					MemoryLimit:   "750Mi",
					CpuRequest:    "100m",
					MemoryRequest: "100Mi",
					Image:         image,
					Requirement: &build.Poste{
						Image:         image,
						GitBranch:     buildVersion,
						VendorVersion: p.PosteVersion,
					},
					Env: []k8s.Env{
						{
							Name:  "HTTPS",
							Value: "OFF",
						},
					},
					Ports: ports,
					Mounts: []deployer.Mount{
						{
							Name:   "poste",
							Target: "/data",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "poste",
					NfsPath:   "/data/poste",
					NfsServer: p.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "poste",
			Name:      "poste",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "poste",
			Name:      "poste",
			Domains:   p.Domains,
		},
	}
}

func (p *Poste) Applier() world.Applier {
	return nil
}

func (p *Poste) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate poste app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate poste app failed")
	}
	if len(p.Domains) == 0 {
		return errors.New("Domains empty")
	}
	if p.PosteVersion == "" {
		return errors.New("PosteVersion empty")
	}
	return nil
}
