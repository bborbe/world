package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type HelloWorld struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
	Tag     docker.Tag
}

func (h *HelloWorld) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *HelloWorld) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
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
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
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
