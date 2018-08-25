package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Mumble struct {
	Cluster cluster.Cluster
	Tag     docker.Tag
}

func (t *Mumble) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Tag,
	)
}

func (m *Mumble) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/mumble",
		Tag:        m.Tag,
	}
	ports := []deployer.Port{
		{
			Name:     "mumble",
			Port:     64738,
			HostPort: 64738,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
		},
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
			Name:      "mumble",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "mumble",
					Image: image,
					Requirement: &build.Mumble{
						Image: image,
					},
					Ports: ports,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "200m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "25Mi",
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
			Name:      "mumble",
			Ports:     ports,
		},
	}
}

func (m *Mumble) Applier() (world.Applier, error) {
	return nil, nil
}
