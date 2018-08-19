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

type Maven struct {
	Cluster          cluster.Cluster
	Domains          []deployer.Domain
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
	nginxImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context: m.Cluster.Context,
			Requirements: []world.Configuration{
				&build.NginxAutoindex{
					Image: nginxImage,
				},
			},
			Namespace: "maven",
			Name:      "api",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "nginx",
					Image:         nginxImage,
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []deployer.Mount{
						{
							Name:     "maven",
							Target:   "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "maven",
					NfsPath:   "/data/maven",
					NfsServer: m.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Domains:   m.Domains,
		},
	}
}

func (m *Maven) api() []world.Configuration {
	mavenRepoImage := docker.Image{
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
			Context: m.Cluster.Context,
			Requirements: []world.Configuration{
				&build.Maven{
					Image: mavenRepoImage,
				},
			},
			Namespace: "maven",
			Name:      "api",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "maven",
					Image:         mavenRepoImage,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=1"},
					Ports:         ports,
					Env: []k8s.Env{
						{
							Name:  "ROOT",
							Value: "/data",
						},
					},
					Mounts: []deployer.Mount{
						{
							Name:   "maven",
							Target: "/data",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "maven",
					NfsPath:   "/data/maven",
					NfsServer: m.Cluster.NfsServer,
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

func (m *Maven) Applier() world.Applier {
	return nil
}

func (m *Maven) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate maven app ...")
	if err := m.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate maven app failed")
	}
	if m.MavenRepoVersion == "" {
		return errors.New("tag missing in maven app")
	}
	if len(m.Domains) == 0 {
		return errors.New("domains empty in maven app")
	}
	return nil
}
