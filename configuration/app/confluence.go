package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Confluence struct {
	Cluster cluster.Cluster
	Domains []world.Domain
	Tag     docker.Tag
}

func (c *Confluence) Childs() []world.Configuration {
	var buildVersion docker.GitBranch = "1.3.0"
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/atlassian-confluence",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", c.Tag, buildVersion)),
	}
	ports := []world.Port{
		{
			Port:     8080,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   c.Cluster.Context,
			Namespace: "confluence",
		},
		&deployer.DeploymentDeployer{
			Context: c.Cluster.Context,
			Requirements: []world.Configuration{
				&build.Confluence{
					VendorVersion: c.Tag,
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
			Context:   c.Cluster.Context,
			Namespace: "confluence",
			Name:      "confluence",
			Ports:     ports,
		},
	}
}

func (c *Confluence) Applier() world.Applier {
	return nil
}

func (c *Confluence) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate confluence app ...")
	if err := c.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate confluence app failed")
	}
	return nil
}
