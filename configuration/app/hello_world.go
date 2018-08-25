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

type HelloWorld struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
	Tag     docker.Tag
}

func (t *HelloWorld) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.Tag,
	)
}

func (h *HelloWorld) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *HelloWorld) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/hello-world",
		Tag:        h.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     80,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
		},
		&deployer.DeploymentDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "hello-world",
					Image: image,
					Requirement: &build.HelloWorld{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Ports: ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Port:      "http",
			Domains:   h.Domains,
		},
	}
}
