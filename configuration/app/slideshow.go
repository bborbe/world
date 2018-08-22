package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Slideshow struct {
	Cluster        cluster.Cluster
	Domains        []k8s.IngressHost
	GitSyncVersion docker.Tag
}

func (s *Slideshow) Applier() (world.Applier, error) {
	return nil, nil
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
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name:     "slideshow",
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
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
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
							Value: "https://github.com/bborbe/slideshow.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/slideshow",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "slideshow",
							Path: "/slideshow",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "slideshow",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
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
			Port:      "http",
			Domains:   s.Domains,
		},
	}
}
