package app

import (
	"fmt"

	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Poste struct {
	Cluster      cluster.Cluster
	PosteVersion docker.Tag
	Domains      k8s.IngressHosts
}

func (t *Poste) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.PosteVersion,
		t.Domains,
	)
}

func (p *Poste) Children() []world.Configuration {
	var buildVersion docker.GitBranch = "1.0.0"
	image := docker.Image{
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
					Name: "poste",
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "1500m",
							Memory: "750Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "100m",
							Memory: "100Mi",
						},
					},
					Image: image,
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
					Mounts: []k8s.VolumeMount{
						{
							Name: "poste",
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "poste",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/poste",
						Server: p.Cluster.NfsServer,
					},
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
			Port:      "http",
			Domains:   p.Domains,
		},
	}
}

func (p *Poste) Applier() (world.Applier, error) {
	return nil, nil
}
