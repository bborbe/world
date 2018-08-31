package app

import (
	"context"

	"github.com/bborbe/world/pkg/world"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Bind struct {
	Cluster cluster.Cluster
	Tag     docker.Tag
}

func (t *Bind) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Tag,
	)
}

func (b *Bind) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/bind",
		Tag:        b.Tag,
	}
	udpPort := deployer.Port{
		Name:     "dns-udp",
		Port:     53,
		HostPort: 53,
		Protocol: "UDP",
	}
	tcpPort := deployer.Port{
		Name:     "dns-tcp",
		Port:     53,
		HostPort: 53,
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "bind",
		},
		&deployer.DeploymentDeployer{
			Context:     b.Cluster.Context,
			Namespace:   "bind",
			Name:        "bind",
			HostNetwork: true,
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "bind",
					Image: image,
					Requirement: &build.Bind{
						Image: image,
					},
					Ports: []deployer.Port{tcpPort, udpPort},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "200m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "25Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "bind",
							Path: "/etc/bind",
						},
						{
							Name: "bind",
							Path: "/var/lib/bind",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: tcpPort.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: tcpPort.Port,
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "bind",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/bind",
						Server: k8s.PodNfsServer(b.Cluster.NfsServer),
					},
				},
			},
		},
	}
}

func (b *Bind) Applier() (world.Applier, error) {
	return nil, nil
}
