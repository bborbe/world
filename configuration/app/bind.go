package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Bind struct {
	Cluster cluster.Cluster
	Tag     docker.Tag
}

func (b *Bind) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/bind",
		Tag:        b.Tag,
	}
	ports := []deployer.Port{
		{
			Name:     "dns-udp",
			Port:     53,
			HostPort: 53,
			Protocol: "UDP",
		},
		{
			Name:     "dns-tcp",
			Port:     53,
			HostPort: 53,
			Protocol: "TCP",
		},
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
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "bind",
					Image: image,
					Requirement: &build.Bind{
						Image: image,
					},
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "200m",
							Memory: "100Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "25Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "bind",
							Path: "/etc/bind",
						},
						{
							Name: "bind",
							Path: "/var/lib/bind",
						},
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
