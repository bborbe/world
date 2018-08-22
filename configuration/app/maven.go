package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Maven struct {
	Cluster          cluster.Cluster
	Domains          []k8s.IngressHost
	MavenRepoVersion docker.Tag
}

func (m *Maven) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
		},
	}
	result = append(result, m.public()...)
	result = append(result, m.api()...)
	return result
}

func (m *Maven) public() []world.Configuration {
	ports := []deployer.Port{
		{
			Port:     80,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "nginx",
					Image: image,
					Requirement: &build.NginxAutoindex{
						Image: image,
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
							Name:     "maven",
							Path:     "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "maven",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/maven",
						Server: m.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
			Port:      "http",
			Domains:   m.Domains,
		},
	}
}

func (m *Maven) api() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/maven-repo",
		Tag:        m.MavenRepoVersion,
	}
	ports := []deployer.Port{
		{
			Port:     8080,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "maven",
					Image: image,
					Requirement: &build.Maven{
						Image: image,
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: ports,
					Env: []k8s.Env{
						{
							Name:  "ROOT",
							Value: "/data",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "maven",
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "maven",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/maven",
						Server: m.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Ports:     ports,
		},
	}
}

func (m *Maven) Applier() (world.Applier, error) {
	return nil, nil
}
