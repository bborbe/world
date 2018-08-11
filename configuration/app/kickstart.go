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

type Kickstart struct {
	Cluster        cluster.Cluster
	Domains        []world.Domain
	GitSyncVersion world.Tag
}

func (k *Kickstart) Applier() world.Applier {
	return nil
}

func (k *Kickstart) Childs() []world.Configuration {
	nginxImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        k.GitSyncVersion,
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
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
		},
		&deployer.DeploymentDeployer{
			Context: k.Cluster.Context,
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
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Name:      "kickstart",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Domains:   k.Domains,
		},
	}
}

func (k *Kickstart) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate kickstart app ...")
	if err := k.Cluster.Validate(ctx); err != nil {
		return err
	}
	if len(k.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	if k.GitSyncVersion == "" {
		return fmt.Errorf("git-sync-version missing")
	}
	return nil
}
