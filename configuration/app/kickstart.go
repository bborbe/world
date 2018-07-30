package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Kickstart struct {
	Context world.Context
	Domains []world.Domain
}

func (d *Kickstart) Applier() world.Applier {
	return nil
}

func (d *Kickstart) Childs() []world.Configuration {
	nginxImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        "1.2.1",
	}
	ports := []world.Port{
		{
			Port: 80,
			Name: "web",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   d.Context,
			Namespace: "kickstart",
		},
		&deployer.DeploymentDeployer{
			Context: d.Context,
			Requirements: []world.Configuration{
				&docker.NginxAutoindex{
					Image: nginxImage,
				},
				&docker.GitSync{
					Image: gitSyncImage,
				},
			},
			Namespace: "kickstart",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "nginx",
					Image:         nginxImage,
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []world.Mount{
						{
							Name:     "kickstart",
							Target:   "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
				{
					Name:          "git-sync",
					Image:         gitSyncImage,
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args: []world.Arg{
						"-logtostderr",
						"-v=4",
					},
					Env: []k8s.Env{
						{
							Name:  "GIT_SYNC_REPO",
							Value: "https://github.com/bborbe/kickstart.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/kickstart",
						},
					},
					Mounts: []world.Mount{
						{
							Name:   "kickstart",
							Target: "/kickstart",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:     "kickstart",
					EmptyDir: true,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "kickstart",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   d.Context,
			Namespace: "kickstart",
			Domains:   d.Domains,
		},
	}
}

func (d *Kickstart) Validate(ctx context.Context) error {
	if d.Context == "" {
		return fmt.Errorf("context missing")
	}
	if len(d.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
