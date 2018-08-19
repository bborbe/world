package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Slideshow struct {
	Cluster        cluster.Cluster
	Domains        []deployer.Domain
	GitSyncVersion docker.Tag
}

func (s *Slideshow) Applier() world.Applier {
	return nil
}

func (s *Slideshow) Children() []world.Configuration {
	nginxImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        s.GitSyncVersion,
	}
	ports := []deployer.Port{
		{
			Port:     80,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
		},
		&deployer.DeploymentDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Name:      "slideshow",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "nginx",
					Image: nginxImage,
					Requirement: &build.NginxAutoindex{
						Image: nginxImage,
					},
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []deployer.Mount{
						{
							Name:     "slideshow",
							Target:   "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
				{
					Name:  "git-sync",
					Image: gitSyncImage,
					Requirement: &build.GitSync{
						Image: gitSyncImage,
					},
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args: []k8s.Arg{
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
					Mounts: []deployer.Mount{
						{
							Name:   "slideshow",
							Target: "/slideshow",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
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
			Name:      "slideshow",
			Domains:   s.Domains,
		},
	}
}

func (s *Slideshow) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate slideshow app ...")
	if err := s.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate slideshow app failed")
	}
	if len(s.Domains) == 0 {
		return errors.New("domains empty in slideshow app")
	}
	if s.GitSyncVersion == "" {
		return errors.New("git-sync-version missing in slideshow app")
	}
	return nil
}
