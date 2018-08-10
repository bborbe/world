package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type HelloWorld struct {
	Context world.Context
	Domains []world.Domain
	Tag     world.Tag
}

func (h *HelloWorld) Applier() world.Applier {
	return nil
}

func (h *HelloWorld) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/hello-world",
		Tag:        h.Tag,
	}
	ports := []world.Port{
		{
			Port:     80,
			Name:     "web",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   h.Context,
			Namespace: "hello-world",
		},
		&deployer.DeploymentDeployer{
			Context: h.Context,
			Requirements: []world.Configuration{
				&docker.HelloWorld{
					Image: image,
				},
			},
			Namespace: "hello-world",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "hello-world",
					Image:         image,
					CpuLimit:      "100",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Ports:         ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   h.Context,
			Namespace: "hello-world",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   h.Context,
			Namespace: "hello-world",
			Domains:   h.Domains,
		},
	}
}

func (h *HelloWorld) Validate(ctx context.Context) error {
	if h.Context == "" {
		return fmt.Errorf("context missing")
	}
	if h.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(h.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
