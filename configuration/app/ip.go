package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type Ip struct {
	Context world.Context
	Domains []world.Domain
	Tag     world.Tag
}

func (i *Ip) Applier() world.Applier {
	return nil
}

func (i *Ip) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
		Tag:        i.Tag,
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
			Context:   i.Context,
			Namespace: "ip",
		},
		&deployer.DeploymentDeployer{
			Context: i.Context,
			Requirements: []world.Configuration{
				&docker.Ip{
					Image: image,
				},
			},
			Namespace: "ip",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "ip",
					Image:         image,
					CpuLimit:      "100",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=2"},
					Ports:         ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   i.Context,
			Namespace: "ip",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   i.Context,
			Namespace: "ip",
			Domains:   i.Domains,
		},
	}
}

func (i *Ip) Validate(ctx context.Context) error {
	if i.Context == "" {
		return fmt.Errorf("context missing")
	}
	if i.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(i.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
