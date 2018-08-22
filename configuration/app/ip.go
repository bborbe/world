package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Ip struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
	Tag     docker.Tag
}

func (i *Ip) Applier() (world.Applier, error) {
	return nil, nil
}

func (i *Ip) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
		Tag:        i.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     8080,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
		},
		&deployer.DeploymentDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "ip",
					Image: image,
					Requirement: &build.Ip{
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
					Args:  []k8s.Arg{"-logtostderr", "-v=2"},
					Ports: ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Port:      "http",
			Domains:   i.Domains,
		},
	}
}
