package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type Bind struct {
	Context   world.Context
	Tag       world.Tag
	NfsServer world.MountNfsServer
}

func (b *Bind) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/bind",
		Tag:        b.Tag,
	}
	ports := []world.Port{
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
			Context:   b.Context,
			Namespace: "bind",
		},
		&deployer.DeploymentDeployer{
			Context: b.Context,
			Requirements: []world.Configuration{
				&docker.Bind{
					Image: image,
				},
			},
			Namespace:   "bind",
			HostNetwork: true,
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "bind",
					Image:         image,
					Ports:         ports,
					CpuLimit:      "200m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "10m",
					MemoryRequest: "25Mi",
					Mounts: []world.Mount{
						{
							Name:   "bind",
							Target: "/etc/bind",
						},
						{
							Name:   "bind",
							Target: "/var/lib/bind",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:      "bind",
					NfsPath:   "/data/bind",
					NfsServer: b.NfsServer,
				},
			},
		},
	}
}

func (b *Bind) Applier() world.Applier {
	return nil
}

func (b *Bind) Validate(ctx context.Context) error {
	if b.Context == "" {
		return fmt.Errorf("context missing")
	}
	if b.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if b.NfsServer == "" {
		return fmt.Errorf("nfs-server missing")
	}
	return nil
}
