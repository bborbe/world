package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/golang/glog"
)

type Password struct {
	Cluster cluster.Cluster
	Domains []world.Domain
	Tag     world.Tag
}

func (p *Password) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/password",
		Tag:        p.Tag,
	}
	ports := []world.Port{
		{
			Port:     8080,
			Name:     "web",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
		},
		&deployer.DeploymentDeployer{
			Context: p.Cluster.Context,
			Requirements: []world.Configuration{
				&docker.Password{
					Image: image,
				},
			},
			Namespace: "password",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "password",
					Image:         image,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=2"},
					Ports:         ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
			Name:      "password",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
			Domains:   p.Domains,
		},
	}
}

func (p *Password) Applier() world.Applier {
	return nil
}

func (p *Password) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate password app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return err
	}
	if p.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(p.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
