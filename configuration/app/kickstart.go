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

type Kickstart struct {
	Cluster        cluster.Cluster
	Domains        []k8s.IngressHost
	GitSyncVersion docker.Tag
}

func (k *Kickstart) Applier() world.Applier {
	return nil
}

func (k *Kickstart) Children() []world.Configuration {
	nginxImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        k.GitSyncVersion,
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
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
		},
		&deployer.DeploymentDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Name:      "kickstart",
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
					Mounts: []k8s.VolumeMount{
						{
							Name:     "kickstart",
							Path:     "/usr/share/nginx/html",
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
							Value: "https://github.com/bborbe/kickstart.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/kickstart",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "kickstart",
							Path: "/kickstart",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "kickstart",
					EmptyDir: &k8s.EmptyDir{},
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
			Name:      "kickstart",
			Port:      "http",
			Domains:   k.Domains,
		},
	}
}

func (k *Kickstart) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate kickstart app ...")
	if err := k.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate kickstart app failed")
	}
	if len(k.Domains) == 0 {
		return errors.New("domains empty in kickstart app")
	}
	if k.GitSyncVersion == "" {
		return errors.New("git-sync-version missing in kickstart app")
	}
	return nil
}
