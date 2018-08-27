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

type Now struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
	Tag     docker.Tag
}

func (t *Now) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.Tag,
	)
}

func (n *Now) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/now",
		Tag:        n.Tag,
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
		},
		&deployer.DeploymentDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "now",
					Image: image,
					Requirement: &build.Now{
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
					Args: []k8s.Arg{"-logtostderr", "-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "8080",
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
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Port:      "http",
			Domains:   n.Domains,
		},
	}
}

func (n *Now) Applier() (world.Applier, error) {
	return nil, nil
}
