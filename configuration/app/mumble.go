package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
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
	port := deployer.Port{
		Name:     "mumble",
		Port:     64738,
		HostPort: 64738,
		Protocol: "TCP",
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
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "mumble",
					Image: image,
					Requirement: &build.Mumble{
						Image: image,
					},
					Ports: []deployer.Port{port},
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
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
			Name:      "mumble",
			Ports:     []deployer.Port{port},
		},
	}
}

func (m *Mumble) Applier() (world.Applier, error) {
	return nil, nil
}