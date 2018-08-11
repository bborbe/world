package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/golang/glog"
)

type Bind struct {
	Cluster cluster.Cluster
	Tag     world.Tag
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
			Context:   b.Cluster.Context,
			Namespace: "bind",
		},
		&deployer.DeploymentDeployer{
			Context: b.Cluster.Context,
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
					NfsServer: b.Cluster.NfsServer,
				},
			},
		},
	}
}

func (b *Bind) Applier() world.Applier {
	return nil
}

func (b *Bind) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate bind app ...")
	if err := b.Cluster.Validate(ctx); err != nil {
		return err
	}
	if b.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	return nil
}
