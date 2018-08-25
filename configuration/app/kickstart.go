package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Kickstart struct {
	Cluster        cluster.Cluster
	Domains        k8s.IngressHosts
	GitSyncVersion docker.Tag
}

func (t *Kickstart) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.GitSyncVersion,
	)
}

func (k *Kickstart) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Kickstart) Children() []world.Configuration {
	nginxImage := docker.Image{
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	gitSyncImage := docker.Image{
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
					Ports: ports,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
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
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
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
					Mounts: []k8s.ContainerMount{
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
					EmptyDir: &k8s.PodVolumeEmptyDir{},
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
