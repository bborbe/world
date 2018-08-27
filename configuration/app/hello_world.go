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
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
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
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
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
					Ports: []deployer.Port{port},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Ports:     []deployer.Port{port},
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
