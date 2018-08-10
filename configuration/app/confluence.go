package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type Confluence struct {
	Context world.Context
	Domains []world.Domain
	Tag     world.Tag
}

func (m *Confluence) Childs() []world.Configuration {
	var buildVersion world.GitBranch = "1.3.0"

	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/atlassian-confluence",
		Tag:        world.Tag(fmt.Sprintf("%s-%s", m.Tag, buildVersion)),
	}
	ports := []world.Port{
		{
			Port:     8080,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Context,
			Namespace: "confluence",
		},
		&deployer.DeploymentDeployer{
			Context: m.Context,
			Requirements: []world.Configuration{
				&docker.Confluence{
					VendorVersion: m.Tag,
					GitBranch:     buildVersion,
					Image:         image,
				},
			},
			Namespace: "confluence",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "confluence",
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
			Namespace: "confluence",
			Ports:     ports,
		},
	}
}

func (m *Confluence) Applier() world.Applier {
	return nil
}

func (m *Confluence) Validate(ctx context.Context) error {
	if m.Context == "" {
		return fmt.Errorf("context missing")
	}
	return nil
}
