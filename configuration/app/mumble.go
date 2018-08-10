package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type Mumble struct {
	Context world.Context
	Tag     world.Tag
}

func (m *Mumble) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/mumble",
		Tag:        m.Tag,
	}
	ports := []world.Port{
		{
			Port:     64738,
			HostPort: 64738,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Context,
			Namespace: "mumble",
		},
		&deployer.DeploymentDeployer{
			Context: m.Context,
			Requirements: []world.Configuration{
				&docker.Mumble{
					Image: image,
				},
			},
			Namespace: "mumble",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "mumble",
					Image:         image,
					Ports:         ports,
					CpuLimit:      "200m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "100m",
					MemoryRequest: "25Mi",
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Context,
			Namespace: "mumble",
			Ports:     ports,
		},
	}
}

func (m *Mumble) Applier() world.Applier {
	return nil
}

func (m *Mumble) Validate(ctx context.Context) error {
	if m.Context == "" {
		return fmt.Errorf("context missing")
	}
	if m.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	return nil
}
