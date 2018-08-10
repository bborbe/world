package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Slideshow struct {
	Context world.Context
	Domains []world.Domain
}

func (d *Slideshow) Applier() world.Applier {
	return nil
}

func (d *Slideshow) Childs() []world.Configuration {
	nginxImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        "1.3.0",
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
			Namespace: "slideshow",
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
			Namespace: "slideshow",
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
							Name:     "slideshow",
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
							Value: "https://github.com/bborbe/slideshow.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/slideshow",
						},
					},
					Mounts: []world.Mount{
						{
							Name:   "slideshow",
							Target: "/slideshow",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:     "slideshow",
					EmptyDir: true,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "slideshow",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   d.Context,
			Namespace: "slideshow",
			Domains:   d.Domains,
		},
	}
}

func (d *Slideshow) Validate(ctx context.Context) error {
	if d.Context == "" {
		return fmt.Errorf("context missing")
	}
	if len(d.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
