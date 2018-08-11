package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

type Slideshow struct {
	Cluster        cluster.Cluster
	Domains        []world.Domain
	GitSyncVersion world.Tag
}

func (s *Slideshow) Applier() world.Applier {
	return nil
}

func (s *Slideshow) Childs() []world.Configuration {
	nginxImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        s.GitSyncVersion,
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
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
		},
		&deployer.DeploymentDeployer{
			Context: s.Cluster.Context,
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
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Name:      "slideshow",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Domains:   s.Domains,
		},
	}
}

func (s *Slideshow) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate slideshow app ...")
	if err := s.Cluster.Validate(ctx); err != nil {
		return err
	}
	if len(s.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	if s.GitSyncVersion == "" {
		return fmt.Errorf("git-sync-version missing")
	}
	return nil
}
