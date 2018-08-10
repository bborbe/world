package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
)

type Download struct {
	Context   world.Context
	Domains   []world.Domain
	NfsServer world.MountNfsServer
}

func (d *Download) Applier() world.Applier {
	return nil
}

func (d *Download) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	ports := []world.Port{
		{
			Port:     80,
			Name:     "web",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   d.Context,
			Namespace: "download",
		},
		&deployer.DeploymentDeployer{
			Context: d.Context,
			Requirements: []world.Configuration{
				&docker.NginxAutoindex{
					Image: image,
				},
			},
			Namespace: "download",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "nginx",
					Image:         image,
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []world.Mount{
						{
							Name:     "download",
							Target:   "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:      "download",
					NfsPath:   "/data/download",
					NfsServer: d.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "download",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   d.Context,
			Namespace: "download",
			Domains:   d.Domains,
		},
	}
}

func (d *Download) Validate(ctx context.Context) error {
	if d.Context == "" {
		return fmt.Errorf("context missing")
	}
	if d.NfsServer == "" {
		return fmt.Errorf("nfs-server missing")
	}
	if len(d.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
