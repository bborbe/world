package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/golang/glog"
	"github.com/pkg/errors"
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
			Context:   b.Cluster.Context,
			Namespace: "bind",
			Name:      "bind",
			Requirements: []world.Configuration{
				&build.Bind{
					Image: image,
				},
			},
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
					Mounts: []deployer.Mount{
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
			Volumes: []deployer.Volume{
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
		return errors.Wrap(err, "validate bind app failed")
	}
	if b.Tag == "" {
		return errors.New("tag missing in bind app")
	}
	return nil
}
