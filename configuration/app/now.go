package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Now struct {
	Context world.Context
	Domains []world.Domain
	Tag     world.Tag
}

func (p *Now) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/now",
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
			Context:   p.Context,
			Namespace: "now",
		},
		&deployer.DeploymentDeployer{
			Context: p.Context,
			Requirements: []world.Configuration{
				&docker.Now{
					Image: image,
				},
			},
			Namespace: "now",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "now",
					Image:         image,
					CpuLimit:      "100",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "8080",
						},
					},
					Ports: ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: "now",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   p.Context,
			Namespace: "now",
			Domains:   p.Domains,
		},
	}
}

func (p *Now) Applier() world.Applier {
	return nil
}

func (p *Now) Validate(ctx context.Context) error {
	if p.Context == "" {
		return fmt.Errorf("context missing")
	}
	if p.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(p.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
